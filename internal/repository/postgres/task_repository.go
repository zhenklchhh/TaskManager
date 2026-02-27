package postgres

import (
	"context"
	"errors"
	"log"
	"log/slog"
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
		INSERT INTO tasks (id, title, type, payload, cron_expr, status, created_at, next_run_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		task.ID, task.Title, task.Type, task.Payload, task.CronExpr, task.Status, task.CreatedAt, task.NextRunAt,
	)
	return err
}

func (r PostgresTaskRepository) GetTaskById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	const q = `
		SELECT id, title, type, payload, cron_expr, status, created_at, retry_count, max_retries,
		last_error_message, updated_at, next_run_at
		FROM tasks
		WHERE id = $1
	`
	var t domain.Task
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.Title, &t.Type, &t.Payload, &t.CronExpr, &t.Status,
		&t.CreatedAt, &t.RetryCount, &t.MaxRetries, &t.LastErrorMsg, &t.UpdatedAt, &t.NextRunAt,
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
        WHERE next_run_at <= NOW() AND status = 'pending'
		ORDER BY next_run_at
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
		FOR UPDATE SKIP LOCKED
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
