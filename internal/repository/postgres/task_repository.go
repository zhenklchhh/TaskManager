package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type PostgresTaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) PostgresTaskRepository {
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
		SELECT id, title, type, payload, cron_expr, created_at, updated_at, next_run_at
		FROM tasks
		WHERE id = $1
	`
	var t domain.Task
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.Title, &t.Type, &t.Payload, &t.CronExpr, &t.CreatedAt, &t.UpdatedAt, &t.NextRunAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	return &t, nil
}
