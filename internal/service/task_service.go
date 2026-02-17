package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type TaskCreateCmd struct {
	Title    string
	Type     string
	Payload  string
	CronExpr string
}

type TaskUpdateStatusCmd struct {
	ID string
	Status domain.TaskStatus
}

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(r repository.TaskRepository) *TaskService {
	return &TaskService{
		repo: r,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, cmd *TaskCreateCmd) (*domain.Task, error) {
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
		Status:     domain.TaskStatusScheduled,
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

func (s *TaskService) GetTaskById(ctx context.Context, id string) (*domain.Task, error) {
	t, err := s.repo.GetTaskById(ctx, id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) GetScheduledTasks(ctx context.Context) ([]string, error) {
	tasks, err := s.repo.GetScheduleTasks(ctx)
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, cmd *TaskUpdateStatusCmd) error {
	id, err := uuid.Parse(cmd.ID)
	if err != nil {
		return err
	}
	err = s.repo.UpdateTaskStatus(ctx, id, string(cmd.Status))
	if err != nil {
		return err
	}
	return err
}