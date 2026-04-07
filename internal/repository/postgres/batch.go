package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

func (r PostgresTaskRepository) BatchCreate(ctx context.Context, tasks []*domain.Task) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("batch create: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	const q = `
		INSERT INTO tasks (id, title, type, payload, cron_expr, status, max_retries, priority, created_at, next_run_at, expires_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	for _, t := range tasks {
		batch.Queue(q, t.ID, t.Title, t.Type, t.Payload, t.CronExpr, t.Status, t.MaxRetries,
			t.Priority, t.CreatedAt, t.NextRunAt, t.ExpiresAt)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	created := 0
	for range tasks {
		_, err := br.Exec()
		if err != nil {
			return created, fmt.Errorf("batch create: exec: %w", err)
		}
		created++
	}

	if err := br.Close(); err != nil {
		return created, fmt.Errorf("batch create: close batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("batch create: commit: %w", err)
	}
	return created, nil
}

func (r PostgresTaskRepository) BatchCancel(ctx context.Context, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	q := fmt.Sprintf(`
		UPDATE tasks
		SET status = 'failed', updated_at = NOW(), last_error_message = 'cancelled by batch operation'
		WHERE id IN (%s) AND status NOT IN ('completed', 'failed')
	`, strings.Join(placeholders, ","))

	res, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("batch cancel: %w", err)
	}
	return int(res.RowsAffected()), nil
}

func (r PostgresTaskRepository) BatchUpdatePriority(ctx context.Context, ids []uuid.UUID, priority int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	args[0] = priority
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	q := fmt.Sprintf(`
		UPDATE tasks
		SET priority = $1, updated_at = NOW()
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	res, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("batch update priority: %w", err)
	}
	return int(res.RowsAffected()), nil
}
