package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type PostgresTaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) repository.TaskRepository {
	return PostgresTaskRepository{pool: pool}
}

func (r PostgresTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	const query = `
		INSERT INTO tasks (id, title, type, payload, cron_expr, status, max_retries, priority, created_at, next_run_at,
		expires_at, company_id, group_id, created_by, assigned_to)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.pool.Exec(ctx, query,
		task.ID, task.Title, task.Type, task.Payload, task.CronExpr, task.Status, task.MaxRetries,
		task.Priority, task.CreatedAt, task.NextRunAt, task.ExpiresAt,
		task.CompanyID, task.GroupID, task.CreatedBy, task.AssignedTo,
	)
	return err
}

func (r PostgresTaskRepository) GetTaskById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	const q = `
		SELECT id, title, type, payload, cron_expr, status, created_at, retry_count, max_retries,
		COALESCE(last_error_message, '') AS last_error_message, updated_at, next_run_at, expires_at, priority,
		company_id, group_id, created_by, assigned_to
		FROM tasks
		WHERE id = $1 AND deleted_at IS NULL
	`
	var t domain.Task
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.Title, &t.Type, &t.Payload, &t.CronExpr, &t.Status,
		&t.CreatedAt, &t.RetryCount, &t.MaxRetries, &t.LastErrorMsg, &t.UpdatedAt, &t.NextRunAt,
		&t.ExpiresAt, &t.Priority,
		&t.CompanyID, &t.GroupID, &t.CreatedBy, &t.AssignedTo,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r PostgresTaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const q = `
		WITH locked_rows AS (
		SELECT id
        FROM tasks
			WHERE next_run_at <= NOW() AND status = 'pending' AND (expires_at IS NULL OR expires_at > NOW()) AND deleted_at IS NULL
		ORDER BY priority ASC, next_run_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
		)
		UPDATE tasks
		SET status = 'scheduled', updated_at = NOW()
		FROM locked_rows
		WHERE tasks.id = locked_rows.id
		RETURNING tasks.id

	`
	tasks := make([]uuid.UUID, 0)
	rows, err := tx.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var nextID uuid.UUID
		if err = rows.Scan(&nextID); err != nil {
			return nil, err
		}
		tasks = append(tasks, nextID)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r PostgresTaskRepository) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error {
	const q = `
		UPDATE tasks
		SET status = $1,
		updated_at = NOW()
		WHERE id = $2
	`
	res, err := r.pool.Exec(ctx, q, status, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		log.Printf("Task %v status didn't updated", id)
	}
	return nil
}

func (r PostgresTaskRepository) UpdateTaskForRetry(ctx context.Context, id uuid.UUID, lastErrorMsg string,
	status domain.TaskStatus, retries int, nextRunAt time.Time) error {
	const q = `
		UPDATE tasks
		SET status = $1, next_run_at = $2, updated_at = NOW(), retry_count = $3, last_error_message = $4
		WHERE id = $5
	`
	res, err := r.pool.Exec(ctx, q, status, nextRunAt, retries, lastErrorMsg, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		slog.Warn("Task have not updated for retry", "id", id)
	}
	return nil
}

func (r PostgresTaskRepository) UpdateStaleTasksToPending(ctx context.Context,
	threshold time.Duration) (int, error) {
	olderThan := time.Now().UTC().Add(-threshold)
	const q = `
		WITH updated_rows AS (
			UPDATE tasks
			SET status = 'pending', updated_at = NOW()
			WHERE updated_at < $1
			  AND status IN ('running', 'scheduled')
			  AND retry_count < max_retries
			  AND deleted_at IS NULL
			RETURNING 1
		)
		SELECT COUNT(*) FROM updated_rows
	`
	var rowsAffected int
	err := r.pool.QueryRow(ctx, q, olderThan).Scan(&rowsAffected)
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (r PostgresTaskRepository) RescheduleTask(ctx context.Context, id uuid.UUID, nextRunAt time.Time) error {
	const q = `
		UPDATE tasks
		SET status = 'pending', next_run_at = $2, retry_count = 0, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.pool.Exec(ctx, q, id, nextRunAt)
	return err
}

func (r PostgresTaskRepository) UpdateTask(ctx context.Context, cmd domain.TaskUpdateCmd) error {
	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argPos := 1

	if cmd.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argPos))
		args = append(args, *cmd.Title)
		argPos++
	}
	if cmd.Payload != nil {
		sets = append(sets, fmt.Sprintf("payload = $%d", argPos))
		args = append(args, []byte(*cmd.Payload))
		argPos++
	}
	if cmd.Priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", argPos))
		args = append(args, *cmd.Priority)
		argPos++
	}
	if cmd.CronExpr != nil {
		sets = append(sets, fmt.Sprintf("cron_expr = $%d", argPos))
		args = append(args, *cmd.CronExpr)
		argPos++
	}
	if cmd.MaxRetries != nil {
		sets = append(sets, fmt.Sprintf("max_retries = $%d", argPos))
		args = append(args, *cmd.MaxRetries)
		argPos++
	}
	if cmd.NextRunAt != nil {
		sets = append(sets, fmt.Sprintf("next_run_at = $%d", argPos))
		args = append(args, *cmd.NextRunAt)
		argPos++
	}
	if cmd.ClearGroup {
		sets = append(sets, "group_id = NULL")
	} else if cmd.GroupID != nil {
		sets = append(sets, fmt.Sprintf("group_id = $%d", argPos))
		args = append(args, *cmd.GroupID)
		argPos++
	}
	if cmd.AssignedTo != nil {
		sets = append(sets, fmt.Sprintf("assigned_to = $%d", argPos))
		args = append(args, *cmd.AssignedTo)
		argPos++
	}

	args = append(args, cmd.ID)
	q := fmt.Sprintf(
		"UPDATE tasks SET %s WHERE id = $%d AND deleted_at IS NULL",
		strings.Join(sets, ", "), argPos,
	)
	res, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r PostgresTaskRepository) SoftDeleteTask(ctx context.Context, id uuid.UUID) error {
	const q = `
		UPDATE tasks
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	res, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r PostgresTaskRepository) CancelTask(ctx context.Context, id uuid.UUID) error {
	const q = `
		UPDATE tasks
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND status IN ('pending', 'running', 'scheduled') AND deleted_at IS NULL
	`
	res, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrTaskNotCancellable
	}
	return nil
}
