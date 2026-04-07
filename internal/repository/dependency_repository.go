package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type DependencyRepository interface {
	CreateDependency(ctx context.Context, dep *domain.TaskDependency) error
	GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskDependency, error)
	GetDependents(ctx context.Context, dependsOnID uuid.UUID) ([]*domain.TaskDependency, error)
	GetChildTasks(ctx context.Context, parentID uuid.UUID) ([]*domain.Task, error)
	CheckAllDependenciesMet(ctx context.Context, taskID uuid.UUID) (bool, error)
	DeleteDependency(ctx context.Context, id uuid.UUID) error
}
