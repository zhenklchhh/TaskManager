package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetTaskById(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	GetPendingTasks(ctx context.Context, limit int) ([]uuid.UUID, error)
	UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error
	UpdateTaskForRetry(ctx context.Context, id uuid.UUID, lastErrorMsg string, status domain.TaskStatus, retries int,
		nextRunAt time.Time) error
	UpdateStaleTasksToPending(ctx context.Context, threshold time.Duration) (int, error)
	GetTaskStats(ctx context.Context) (*domain.TaskStats, error)
	GetAllTasks(ctx context.Context, limit, offset int, status *domain.TaskStatus) ([]*domain.Task, error)
	GetTaskCount(ctx context.Context, status *domain.TaskStatus) (int, error)
}
