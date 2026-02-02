package repository

import (
	"context"

	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetTaskById(ctx context.Context, id string) (*domain.Task, error)
}
