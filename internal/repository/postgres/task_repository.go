package postgres

import (
	"context"
	"errors"
	"log"

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

func (r PostgresTaskRepository) GetTaskById(ctx context.Context, id string) (*domain.Task, error) {
	const q = `
		SELECT id, title, type, payload, cron_expr, status, retry_count, max_retries, created_at, updated_at, next_run_at
		FROM tasks
		WHERE id = $1
	`
	var t domain.Task
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.Title, &t.Type, &t.Payload, &t.CronExpr, &t.Status, &t.RetryCount,
		&t.MaxRetries, &t.CreatedAt, &t.UpdatedAt, &t.NextRunAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	return &t, nil
}
// 2026/02/18 16:03:29 scheduler: error while checking upcoming tasks: can't scan into dest[0] (col: id):
//  cannot scan uuid (OID 2950) in binary format into []string

func (r PostgresTaskRepository) GetScheduleTasks(ctx context.Context) ([]uuid.UUID, error) {
	const q = `
		SELECT id
        FROM tasks
        WHERE next_run_at <= NOW() AND status = 'scheduled'
	`
	tasks := make([]uuid.UUID, 0)
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		// cant't scan uuid into string, maybe use pgtype.UUID
		var stringID string
		err = rows.Scan(stringID)
		if err != nil {
			return nil, err
		}
		id, err := uuid.Parse(stringID)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, id)
	}
	return tasks, err
}

func (r PostgresTaskRepository) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status string) error {
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
