package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type PostgresExecutionRepository struct {
	pool *pgxpool.Pool
}

func NewExecutionRepository(pool *pgxpool.Pool) repository.ExecutionRepository {
	return PostgresExecutionRepository{pool: pool}
}

func (r PostgresExecutionRepository) CreateExecution(ctx context.Context, exec *domain.TaskRun) error {
	const q = `
		INSERT INTO task_executions (id, task_id, started_at, status, worker_id)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, q, exec.ID, exec.TaskID, exec.StartedAt, string(exec.Status), exec.WorkerID)
	return err
}

func (r PostgresExecutionRepository) FinishExecution(
	ctx context.Context,
	id uuid.UUID,
	status domain.TaskStatus,
	errMsg, output string,
	durationMs int64,
) error {
	const q = `
		UPDATE task_executions
		SET finished_at = NOW(),
		    status = $2,
		    error_message = NULLIF($3, ''),
		    output = NULLIF($4, ''),
		    duration_ms = $5
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, q, id, string(status), errMsg, output, durationMs)
	return err
}

func (r PostgresExecutionRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskRun, error) {
	const q = `
		SELECT id, task_id, started_at, finished_at, status,
		       COALESCE(error_message, ''),
		       COALESCE(output, ''),
		       COALESCE(worker_id, ''),
		       COALESCE(duration_ms, 0)
		FROM task_executions
		WHERE task_id = $1
		ORDER BY started_at DESC
		LIMIT 50
	`
	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("get executions by task_id: %w", err)
	}
	defer rows.Close()

	result := make([]*domain.TaskRun, 0)
	for rows.Next() {
		var run domain.TaskRun
		var statusStr string
		if err := rows.Scan(
			&run.ID, &run.TaskID, &run.StartedAt, &run.FinishedAt,
			&statusStr, &run.Error, &run.Output, &run.WorkerID, &run.DurationMs,
		); err != nil {
			return nil, fmt.Errorf("get executions by task_id: scan: %w", err)
		}
		run.Status = domain.TaskStatus(statusStr)
		result = append(result, &run)
	}
	return result, nil
}
