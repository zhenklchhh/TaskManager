package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type PostgresDependencyRepository struct {
	pool *pgxpool.Pool
}

func NewDependencyRepository(pool *pgxpool.Pool) repository.DependencyRepository {
	return &PostgresDependencyRepository{pool: pool}
}

func (r *PostgresDependencyRepository) CreateDependency(ctx context.Context, dep *domain.TaskDependency) error {
	const q = `
		INSERT INTO task_dependencies (id, task_id, depends_on_id, condition, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, q, dep.ID, dep.TaskID, dep.DependsOnID, dep.Condition, dep.CreatedAt)
	if err != nil {
		return fmt.Errorf("create dependency: %w", err)
	}
	return nil
}

func (r *PostgresDependencyRepository) GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskDependency, error) {
	const q = `
		SELECT id, task_id, depends_on_id, condition, created_at
		FROM task_dependencies
		WHERE task_id = $1
	`
	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("get dependencies: %w", err)
	}
	defer rows.Close()

	deps := make([]*domain.TaskDependency, 0)
	for rows.Next() {
		var d domain.TaskDependency
		if err := rows.Scan(&d.ID, &d.TaskID, &d.DependsOnID, &d.Condition, &d.CreatedAt); err != nil {
			return nil, err
		}
		deps = append(deps, &d)
	}
	return deps, nil
}

func (r *PostgresDependencyRepository) GetDependents(ctx context.Context, dependsOnID uuid.UUID) ([]*domain.TaskDependency, error) {
	const q = `
		SELECT id, task_id, depends_on_id, condition, created_at
		FROM task_dependencies
		WHERE depends_on_id = $1
	`
	rows, err := r.pool.Query(ctx, q, dependsOnID)
	if err != nil {
		return nil, fmt.Errorf("get dependents: %w", err)
	}
	defer rows.Close()

	deps := make([]*domain.TaskDependency, 0)
	for rows.Next() {
		var d domain.TaskDependency
		if err := rows.Scan(&d.ID, &d.TaskID, &d.DependsOnID, &d.Condition, &d.CreatedAt); err != nil {
			return nil, err
		}
		deps = append(deps, &d)
	}
	return deps, nil
}

func (r *PostgresDependencyRepository) GetChildTasks(ctx context.Context, parentID uuid.UUID) ([]*domain.Task, error) {
	const q = `
		SELECT id, title, type, payload, cron_expr, status, created_at, retry_count, max_retries,
		COALESCE(last_error_message, ''), updated_at, next_run_at, expires_at, priority, parent_id
		FROM tasks
		WHERE parent_id = $1
	`
	rows, err := r.pool.Query(ctx, q, parentID)
	if err != nil {
		return nil, fmt.Errorf("get child tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Type, &t.Payload, &t.CronExpr, &t.Status,
			&t.CreatedAt, &t.RetryCount, &t.MaxRetries, &t.LastErrorMsg, &t.UpdatedAt,
			&t.NextRunAt, &t.ExpiresAt, &t.Priority, &t.ParentID); err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *PostgresDependencyRepository) CheckAllDependenciesMet(ctx context.Context, taskID uuid.UUID) (bool, error) {
	const q = `
		SELECT COUNT(*) FROM task_dependencies td
		JOIN tasks t ON t.id = td.depends_on_id
		WHERE td.task_id = $1
		AND NOT (
			(td.condition = 'completed' AND t.status = 'completed')
			OR (td.condition = 'failed' AND t.status = 'failed')
			OR (td.condition = 'any' AND t.status IN ('completed', 'failed'))
		)
	`
	var unmetCount int
	err := r.pool.QueryRow(ctx, q, taskID).Scan(&unmetCount)
	if err != nil {
		return false, fmt.Errorf("check dependencies met: %w", err)
	}
	return unmetCount == 0, nil
}

func (r *PostgresDependencyRepository) DeleteDependency(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM task_dependencies WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete dependency: %w", err)
	}
	return nil
}
