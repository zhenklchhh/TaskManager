package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetTaskById(ctx context.Context, id string) (*domain.Task, error)
	GetPendingTasks(ctx context.Context) ([]uuid.UUID, error)
	UpdateTaskStatus(ctx context.Context, id uuid.UUID, status string) error
}
