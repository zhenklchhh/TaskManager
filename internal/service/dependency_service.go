package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type DependencyService struct {
	depRepo  repository.DependencyRepository
	taskRepo repository.TaskRepository
}

func NewDependencyService(depRepo repository.DependencyRepository, taskRepo repository.TaskRepository) *DependencyService {
	return &DependencyService{
		depRepo:  depRepo,
		taskRepo: taskRepo,
	}
}

func (s *DependencyService) AddDependency(ctx context.Context, taskID, dependsOnID uuid.UUID, condition domain.DependencyCondition) (*domain.TaskDependency, error) {
	if taskID == dependsOnID {
		return nil, domain.ErrValidation
	}

	if condition == "" {
		condition = domain.ConditionCompleted
	}

	dep := &domain.TaskDependency{
		ID:          uuid.New(),
		TaskID:      taskID,
		DependsOnID: dependsOnID,
		Condition:   condition,
		CreatedAt:   time.Now(),
	}

	if err := s.depRepo.CreateDependency(ctx, dep); err != nil {
		return nil, err
	}
	return dep, nil
}

func (s *DependencyService) GetDependencies(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskDependency, error) {
	return s.depRepo.GetDependencies(ctx, taskID)
}

func (s *DependencyService) GetDependents(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskDependency, error) {
	return s.depRepo.GetDependents(ctx, taskID)
}

func (s *DependencyService) GetChildTasks(ctx context.Context, parentID uuid.UUID) ([]*domain.Task, error) {
	return s.depRepo.GetChildTasks(ctx, parentID)
}

func (s *DependencyService) RemoveDependency(ctx context.Context, id uuid.UUID) error {
	return s.depRepo.DeleteDependency(ctx, id)
}

func (s *DependencyService) OnTaskCompleted(ctx context.Context, completedTaskID uuid.UUID, completedStatus domain.TaskStatus) {
	dependents, err := s.depRepo.GetDependents(ctx, completedTaskID)
	if err != nil {
		slog.Error("dependency: failed to get dependents", "task_id", completedTaskID, "error", err)
		return
	}

	for _, dep := range dependents {
		conditionMet := false
		switch dep.Condition {
		case domain.ConditionCompleted:
			conditionMet = completedStatus == domain.TaskStatusCompleted
		case domain.ConditionFailed:
			conditionMet = completedStatus == domain.TaskStatusFailed
		case domain.ConditionAny:
			conditionMet = completedStatus == domain.TaskStatusCompleted || completedStatus == domain.TaskStatusFailed
		}

		if !conditionMet {
			continue
		}

		allMet, err := s.depRepo.CheckAllDependenciesMet(ctx, dep.TaskID)
		if err != nil {
			slog.Error("dependency: failed to check dependencies", "task_id", dep.TaskID, "error", err)
			continue
		}

		if allMet {
			slog.Info("dependency: all dependencies met, activating task", "task_id", dep.TaskID)
			if err := s.taskRepo.UpdateTaskStatus(ctx, dep.TaskID, domain.TaskStatusPending); err != nil {
				slog.Error("dependency: failed to activate dependent task", "task_id", dep.TaskID, "error", err)
			}
		}
	}

	children, err := s.depRepo.GetChildTasks(ctx, completedTaskID)
	if err != nil {
		slog.Error("dependency: failed to get child tasks", "parent_id", completedTaskID, "error", err)
		return
	}

	for _, child := range children {
		if child.Status == domain.TaskStatusPending {
			continue
		}
		allMet, err := s.depRepo.CheckAllDependenciesMet(ctx, child.ID)
		if err != nil {
			slog.Error("dependency: failed to check child dependencies", "child_id", child.ID, "error", err)
			continue
		}
		if allMet {
			slog.Info("dependency: activating child task", "child_id", child.ID, "parent_id", completedTaskID)
			if err := s.taskRepo.UpdateTaskStatus(ctx, child.ID, domain.TaskStatusPending); err != nil {
				slog.Error("dependency: failed to activate child task", "child_id", child.ID, "error", err)
			}
		}
	}
}
