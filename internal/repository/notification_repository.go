package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type NotificationRepository interface {
	CreateConfig(ctx context.Context, cfg *domain.NotificationConfig) error
	GetConfigsByTaskAndEvent(ctx context.Context, taskID uuid.UUID, event domain.NotificationEvent) ([]*domain.NotificationConfig, error)
	GetGlobalConfigsByEvent(ctx context.Context, event domain.NotificationEvent) ([]*domain.NotificationConfig, error)
	CreateLog(ctx context.Context, log *domain.NotificationLog) error
	GetPendingLogs(ctx context.Context, limit int) ([]*domain.NotificationLog, error)
	UpdateLog(ctx context.Context, log *domain.NotificationLog) error
	GetConfigByID(ctx context.Context, id uuid.UUID) (*domain.NotificationConfig, error)
}
