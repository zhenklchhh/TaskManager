package service

import (
	"context"
	"log/slog"
	"math"
	"math/rand"
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
	repo                  repository.TaskRepository
	defaultTaskMaxRetries int
}

func NewTaskService(r repository.TaskRepository, maxRetries int) *TaskService {
	return &TaskService{
		repo:                  r,
		defaultTaskMaxRetries: maxRetries,
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

	var (
		maxRetries int
		priority   int
		expiresAt  *time.Time
	)

	if cmd.MaxRetries == nil {
		maxRetries = s.defaultTaskMaxRetries
	} else {
		maxRetries = *cmd.MaxRetries
	}

	if cmd.Priority == nil {
		priority = 5
	} else {
		priority = *cmd.Priority
		if priority < 1 {
			priority = 1
		}
		if priority > 10 {
			priority = 10
		}
	}

	if cmd.ExpiresAt != nil {
		expiresAt = cmd.ExpiresAt
	}

	now := time.Now()
	nextAt := sch.Next(now)

	id := uuid.New()
	t := &domain.Task{
		ID:         id,
		Title:      cmd.Title,
		Type:       cmd.Type,
		Payload:    []byte(cmd.Payload),
		CronExpr:   cmd.CronExpr,
		Status:     domain.TaskStatusPending,
		RetryCount: 0,
		MaxRetries: maxRetries,
		Priority:   priority,
		CreatedAt:  now,
		UpdatedAt:  now,
		NextRunAt:  nextAt,
		ExpiresAt:  expiresAt,
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
		slog.Error(
			"Task failed after reaching max retries", "id", id.String(),
			"retries", task.MaxRetries,
			"error", taskError.Error(),
		)
		return s.UpdateTaskStatus(ctx, &domain.TaskUpdateStatusCmd{
			ID:           id,
			Status:       domain.TaskStatusFailed,
			LastErrorMsg: domain.ErrMaxRetriesExceeded.Error(),
		})
	}
	newRetriesCount := task.RetryCount + 1

	backoffSeconds := math.Pow(2, float64(newRetriesCount)) * 60
	const maxBackoffSeconds = 3600
	if backoffSeconds > maxBackoffSeconds {
		backoffSeconds = maxBackoffSeconds
	}
	base := int64(backoffSeconds / 2)
	if base < 1 {
		base = 1
	}
	jitter := rand.Int63n(base)
	finalSeconds := base + jitter
	nextRunAt := time.Now().UTC().Add(time.Duration(finalSeconds) * time.Second)
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

func (s *TaskService) GetTaskStats(ctx context.Context) (*domain.TaskStats, error) {
	return s.repo.GetTaskStats(ctx)
}

func (s *TaskService) GetAllTasks(ctx context.Context, limit, offset int, status *domain.TaskStatus) ([]*domain.Task, error) {
	return s.repo.GetAllTasks(ctx, limit, offset, status)
}

func (s *TaskService) GetTaskCount(ctx context.Context, status *domain.TaskStatus) (int, error) {
	return s.repo.GetTaskCount(ctx, status)
}
