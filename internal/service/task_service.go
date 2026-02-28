package service

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type TaskServiceInterface interface {
	CreateTask(ctx context.Context, cmd *domain.TaskCreateCmd) (*domain.Task, error)
	GetTaskById(ctx context.Context, id uuid.UUID) (*domain.Task, error)
}

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(r repository.TaskRepository) *TaskService {
	return &TaskService{
		repo: r,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, cmd *domain.TaskCreateCmd) (*domain.Task, error) {
	if cmd.CronExpr == "" || cmd.Type == "" || cmd.Title == "" || cmd.Payload == "" {
		return nil, domain.ErrValidation
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sch, err := parser.Parse(cmd.CronExpr)
	if err != nil {
		return nil, domain.ErrInvalidCron
	}
	now := time.Now()
	nextAt := sch.Next(now)

	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	t := &domain.Task{
		ID:         uuid,
		Title:      cmd.Title,
		Type:       cmd.Type,
		Payload:    []byte(cmd.Payload),
		CronExpr:   cmd.CronExpr,
		Status:     domain.TaskStatusPending,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  now,
		UpdatedAt:  now,
		NextRunAt:  nextAt,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) GetTaskById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	t, err := s.repo.GetTaskById(ctx, id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) ProcessPendingTasks(ctx context.Context, limit int) ([]uuid.UUID, error) {
	tasks, err := s.repo.GetPendingTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, cmd *domain.TaskUpdateStatusCmd) error {
	return s.repo.UpdateTaskStatus(ctx, cmd.ID, cmd.Status)
}

func (s *TaskService) UpdateTaskForRetry(ctx context.Context, cmd *domain.TaskUpdateForRetryCmd) error {
	return s.repo.UpdateTaskForRetry(ctx, cmd.ID, cmd.LastErrorMsg, cmd.Status, cmd.Retries, cmd.NextRunAt)
}

func (s *TaskService) RetryTask(ctx context.Context, id uuid.UUID, taskError error) error {
	task, err := s.GetTaskById(ctx, id)
	if err != nil {
		return err
	}
	if task.RetryCount >= task.MaxRetries {
		return s.UpdateTaskStatus(ctx, &domain.TaskUpdateStatusCmd{
			ID:     id,
			Status: domain.TaskStatusFailed,
		})
	}

	newRetriesCount := task.RetryCount + 1
	backoff := time.Duration(math.Pow(2, float64(newRetriesCount))) * time.Minute
	nextRunAt := time.Now().UTC().Add(backoff)
	return s.UpdateTaskForRetry(ctx, &domain.TaskUpdateForRetryCmd{
		ID:           id,
		Status:       domain.TaskStatusPending,
		LastErrorMsg: taskError.Error(),
		Retries:      newRetriesCount,
		NextRunAt:    nextRunAt,
	})
}

func (s *TaskService) UpdateStaleTasksToPending(ctx context.Context, threshold time.Duration) (int, error) {
	return s.repo.UpdateStaleTasksToPending(ctx, threshold)
}
