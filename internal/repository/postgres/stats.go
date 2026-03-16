package postgres

import (
	"context"

	"github.com/zhenklchhh/TaskManager/internal/domain"
)

func (r PostgresTaskRepository) GetTaskStats(ctx context.Context) (*domain.TaskStats, error) {
	stats := &domain.TaskStats{
		TasksByType:     make(map[string]int),
		TasksByPriority: make(map[int]int),
	}

	const q = `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'scheduled') as scheduled,
			COUNT(*) FILTER (WHERE status = 'running') as running,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		FROM tasks
	`
	err := r.pool.QueryRow(ctx, q).Scan(
		&stats.TotalTasks, &stats.PendingTasks, &stats.ScheduledTasks,
		&stats.RunningTasks, &stats.CompletedTasks, &stats.FailedTasks,
	)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `SELECT type, COUNT(*) FROM tasks GROUP BY type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var taskType string
		var count int
		if err := rows.Scan(&taskType, &count); err != nil {
			return nil, err
		}
		stats.TasksByType[taskType] = count
	}

	rows2, err := r.pool.Query(ctx, `SELECT priority, COUNT(*) FROM tasks GROUP BY priority`)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var priority, count int
		if err := rows2.Scan(&priority, &count); err != nil {
			return nil, err
		}
		stats.TasksByPriority[priority] = count
	}

	return stats, nil
}

func (r PostgresTaskRepository) GetAllTasks(ctx context.Context, limit, offset int, status *domain.TaskStatus) ([]*domain.Task, error) {
	query := `
		SELECT id, title, type, payload, cron_expr, status, created_at, retry_count, max_retries,
		COALESCE(last_error_message, ''), updated_at, next_run_at, expires_at, priority
		FROM tasks
	`
	args := []interface{}{}
	argPos := 1

	if status != nil {
		query += ` WHERE status = $` + string(rune(argPos+'0'))
		args = append(args, *status)
		argPos++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + string(rune(argPos+'0')) + ` OFFSET $` + string(rune(argPos+'1'))
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		var t domain.Task
		err := rows.Scan(&t.ID, &t.Title, &t.Type, &t.Payload, &t.CronExpr, &t.Status,
			&t.CreatedAt, &t.RetryCount, &t.MaxRetries, &t.LastErrorMsg, &t.UpdatedAt,
			&t.NextRunAt, &t.ExpiresAt, &t.Priority)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}

	return tasks, nil
}

func (r PostgresTaskRepository) GetTaskCount(ctx context.Context, status *domain.TaskStatus) (int, error) {
	query := `SELECT COUNT(*) FROM tasks`
	args := []interface{}{}

	if status != nil {
		query += ` WHERE status = $1`
		args = append(args, *status)
	}

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}
