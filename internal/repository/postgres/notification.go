package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type PostgresNotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) repository.NotificationRepository {
	return &PostgresNotificationRepository{pool: pool}
}

func (r *PostgresNotificationRepository) CreateConfig(ctx context.Context, cfg *domain.NotificationConfig) error {
	const q = `
		INSERT INTO notification_configs (id, task_id, type, event, target, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, q, cfg.ID, cfg.TaskID, cfg.Type, cfg.Event, cfg.Target, cfg.CreatedAt)
	if err != nil {
		return fmt.Errorf("create notification config: %w", err)
	}
	return nil
}

func (r *PostgresNotificationRepository) GetConfigsByTaskAndEvent(ctx context.Context, taskID uuid.UUID, event domain.NotificationEvent) ([]*domain.NotificationConfig, error) {
	const q = `
		SELECT id, task_id, type, event, target, created_at
		FROM notification_configs
		WHERE task_id = $1 AND event = $2
	`
	rows, err := r.pool.Query(ctx, q, taskID, event)
	if err != nil {
		return nil, fmt.Errorf("get configs by task and event: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

func (r *PostgresNotificationRepository) GetGlobalConfigsByEvent(ctx context.Context, event domain.NotificationEvent) ([]*domain.NotificationConfig, error) {
	const q = `
		SELECT id, task_id, type, event, target, created_at
		FROM notification_configs
		WHERE task_id IS NULL AND event = $1
	`
	rows, err := r.pool.Query(ctx, q, event)
	if err != nil {
		return nil, fmt.Errorf("get global configs by event: %w", err)
	}
	defer rows.Close()
	return scanConfigs(rows)
}

func scanConfigs(rows pgx.Rows) ([]*domain.NotificationConfig, error) {
	configs := make([]*domain.NotificationConfig, 0)
	for rows.Next() {
		var c domain.NotificationConfig
		if err := rows.Scan(&c.ID, &c.TaskID, &c.Type, &c.Event, &c.Target, &c.CreatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, &c)
	}
	return configs, nil
}

func (r *PostgresNotificationRepository) CreateLog(ctx context.Context, log *domain.NotificationLog) error {
	const q = `
		INSERT INTO notification_logs (id, config_id, task_id, event, status, attempts, max_attempts, last_error, next_retry_at, created_at, last_attempt_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, q, log.ID, log.ConfigID, log.TaskID, log.Event, log.Status,
		log.Attempts, log.MaxAttempts, log.LastError, log.NextRetryAt, log.CreatedAt, log.LastAttemptAt)
	if err != nil {
		return fmt.Errorf("create notification log: %w", err)
	}
	return nil
}

func (r *PostgresNotificationRepository) GetPendingLogs(ctx context.Context, limit int) ([]*domain.NotificationLog, error) {
	const q = `
		SELECT id, config_id, task_id, event, status, attempts, max_attempts,
			COALESCE(last_error, ''), next_retry_at, created_at, last_attempt_at
		FROM notification_logs
		WHERE status = 'pending' AND (next_retry_at IS NULL OR next_retry_at <= $1)
		ORDER BY created_at ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, q, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("get pending logs: %w", err)
	}
	defer rows.Close()

	logs := make([]*domain.NotificationLog, 0)
	for rows.Next() {
		var l domain.NotificationLog
		if err := rows.Scan(&l.ID, &l.ConfigID, &l.TaskID, &l.Event, &l.Status,
			&l.Attempts, &l.MaxAttempts, &l.LastError, &l.NextRetryAt, &l.CreatedAt, &l.LastAttemptAt); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

func (r *PostgresNotificationRepository) UpdateLog(ctx context.Context, log *domain.NotificationLog) error {
	const q = `
		UPDATE notification_logs
		SET status = $1, attempts = $2, last_error = $3, next_retry_at = $4, last_attempt_at = $5
		WHERE id = $6
	`
	_, err := r.pool.Exec(ctx, q, log.Status, log.Attempts, log.LastError, log.NextRetryAt, log.LastAttemptAt, log.ID)
	if err != nil {
		return fmt.Errorf("update notification log: %w", err)
	}
	return nil
}

func (r *PostgresNotificationRepository) GetConfigByID(ctx context.Context, id uuid.UUID) (*domain.NotificationConfig, error) {
	const q = `
		SELECT id, task_id, type, event, target, created_at
		FROM notification_configs
		WHERE id = $1
	`
	var c domain.NotificationConfig
	err := r.pool.QueryRow(ctx, q, id).Scan(&c.ID, &c.TaskID, &c.Type, &c.Event, &c.Target, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("notification config not found")
		}
		return nil, err
	}
	return &c, nil
}
