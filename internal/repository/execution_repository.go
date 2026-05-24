package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type ExecutionRepository interface {
	CreateExecution(ctx context.Context, exec *domain.TaskRun) error
	FinishExecution(ctx context.Context, id uuid.UUID, status domain.TaskStatus, errMsg, output string, durationMs int64) error
	GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskRun, error)
}
