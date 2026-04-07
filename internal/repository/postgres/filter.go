package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/zhenklchhh/TaskManager/internal/domain"
)

func buildFilterQuery(base string, filter domain.TaskFilter) (string, []interface{}) {
	conditions := []string{}
	args := []interface{}{}
	argPos := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
		args = append(args, string(*filter.Status))
		argPos++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argPos))
		args = append(args, *filter.Type)
		argPos++
	}
	if filter.PriorityMin != nil {
		conditions = append(conditions, fmt.Sprintf("priority >= $%d", argPos))
		args = append(args, *filter.PriorityMin)
		argPos++
	}
	if filter.PriorityMax != nil {
		conditions = append(conditions, fmt.Sprintf("priority <= $%d", argPos))
		args = append(args, *filter.PriorityMax)
		argPos++
	}
	if filter.CreatedFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argPos))
		args = append(args, *filter.CreatedFrom)
		argPos++
	}
	if filter.CreatedTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argPos))
		args = append(args, *filter.CreatedTo)
		argPos++
	}

	query := base
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}

func (r PostgresTaskRepository) GetAllTasksFiltered(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, error) {
	base := `
		SELECT id, title, type, payload, cron_expr, status, created_at, retry_count, max_retries,
		COALESCE(last_error_message, ''), updated_at, next_run_at, expires_at, priority
		FROM tasks
	`
	query, args := buildFilterQuery(base, filter)
	argPos := len(args) + 1

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get all tasks filtered: %w", err)
	}
	defer rows.Close()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		var t domain.Task
		err := rows.Scan(&t.ID, &t.Title, &t.Type, &t.Payload, &t.CronExpr, &t.Status,
			&t.CreatedAt, &t.RetryCount, &t.MaxRetries, &t.LastErrorMsg, &t.UpdatedAt,
			&t.NextRunAt, &t.ExpiresAt, &t.Priority)
		if err != nil {
			return nil, fmt.Errorf("get all tasks filtered: scan: %w", err)
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r PostgresTaskRepository) GetTaskCountFiltered(ctx context.Context, filter domain.TaskFilter) (int, error) {
	base := `SELECT COUNT(*) FROM tasks`
	query, args := buildFilterQuery(base, filter)

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get task count filtered: %w", err)
	}
	return count, err
}
