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
	UpdateTask(ctx context.Context, cmd *domain.TaskUpdateCmd) (*domain.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	CancelTask(ctx context.Context, id uuid.UUID) error
	ManualRetry(ctx context.Context, id uuid.UUID) error
	GetTaskExecutions(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskRun, error)
}

type TaskService struct {
	repo                  repository.TaskRepository
	execRepo              repository.ExecutionRepository
	defaultTaskMaxRetries int
}

func NewTaskService(r repository.TaskRepository, maxRetries int) *TaskService {
	return &TaskService{
		repo:                  r,
		defaultTaskMaxRetries: maxRetries,
	}
}

func (s *TaskService) SetExecutionRepository(r repository.ExecutionRepository) {
	s.execRepo = r
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
		CompanyID:  cmd.CompanyID,
		GroupID:    cmd.GroupID,
		CreatedBy:  cmd.CreatedBy,
		AssignedTo: cmd.AssignedTo,
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

func (s *TaskService) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	task, err := s.repo.GetTaskById(ctx, id)
	if err != nil {
		return err
	}
	if task.CronExpr != "" {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		sch, err := parser.Parse(task.CronExpr)
		if err != nil {
			slog.Warn("MarkCompleted: invalid cron_expr, marking as completed", "id", id, "cron", task.CronExpr)
			return s.repo.UpdateTaskStatus(ctx, id, domain.TaskStatusCompleted)
		}
		nextAt := sch.Next(time.Now())
		return s.repo.RescheduleTask(ctx, id, nextAt)
	}
	return s.repo.UpdateTaskStatus(ctx, id, domain.TaskStatusCompleted)
}

func (s *TaskService) UpdateTask(ctx context.Context, cmd *domain.TaskUpdateCmd) (*domain.Task, error) {
	if cmd.CronExpr != nil {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		sch, err := parser.Parse(*cmd.CronExpr)
		if err != nil {
			return nil, domain.ErrInvalidCron
		}
		nextAt := sch.Next(time.Now())
		cmd.NextRunAt = &nextAt
	}
	if cmd.Priority != nil {
		if *cmd.Priority < 1 {
			*cmd.Priority = 1
		}
		if *cmd.Priority > 10 {
			*cmd.Priority = 10
		}
	}
	if err := s.repo.UpdateTask(ctx, *cmd); err != nil {
		return nil, err
	}
	return s.repo.GetTaskById(ctx, cmd.ID)
}

func (s *TaskService) DeleteTask(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetTaskById(ctx, id); err != nil {
		return err
	}
	return s.repo.SoftDeleteTask(ctx, id)
}

func (s *TaskService) CancelTask(ctx context.Context, id uuid.UUID) error {
	task, err := s.repo.GetTaskById(ctx, id)
	if err != nil {
		return err
	}
	switch task.Status {
	case domain.TaskStatusCompleted, domain.TaskStatusFailed, domain.TaskStatusCancelled:
		return domain.ErrTaskNotCancellable
	}
	return s.repo.CancelTask(ctx, id)
}

func (s *TaskService) ManualRetry(ctx context.Context, id uuid.UUID) error {
	task, err := s.repo.GetTaskById(ctx, id)
	if err != nil {
		return err
	}
	if task.Status != domain.TaskStatusFailed && task.Status != domain.TaskStatusCancelled {
		return domain.ErrValidation
	}
	return s.repo.UpdateTaskForRetry(ctx, id, "", domain.TaskStatusPending, 0, time.Now().UTC())
}

func (s *TaskService) CreateExecution(ctx context.Context, taskID uuid.UUID, workerID string) (*domain.TaskRun, error) {
	run := &domain.TaskRun{
		ID:        uuid.New(),
		TaskID:    taskID,
		StartedAt: time.Now().UTC(),
		Status:    domain.TaskStatusRunning,
		WorkerID:  workerID,
	}
	if s.execRepo != nil {
		if err := s.execRepo.CreateExecution(ctx, run); err != nil {
			slog.Warn("CreateExecution: failed to persist", "task_id", taskID, "error", err)
		}
	}
	return run, nil
}

func (s *TaskService) FinishExecution(ctx context.Context, id uuid.UUID, status domain.TaskStatus, errMsg, output string, durationMs int64) error {
	if s.execRepo == nil {
		return nil
	}
	return s.execRepo.FinishExecution(ctx, id, status, errMsg, output, durationMs)
}

func (s *TaskService) GetTaskExecutions(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskRun, error) {
	if s.execRepo == nil {
		return []*domain.TaskRun{}, nil
	}
	return s.execRepo.GetByTaskID(ctx, taskID)
}

func (s *TaskService) GetTaskStats(ctx context.Context) (*domain.TaskStats, error) {
	return s.repo.GetTaskStats(ctx)
}

func (s *TaskService) GetTaskStatsForCompany(ctx context.Context, companyID uuid.UUID) (*domain.TaskStats, error) {
	return s.repo.GetTaskStatsForCompany(ctx, companyID)
}

func (s *TaskService) GetAllTasks(ctx context.Context, limit, offset int, status *domain.TaskStatus) ([]*domain.Task, error) {
	return s.repo.GetAllTasks(ctx, limit, offset, status)
}

func (s *TaskService) GetTaskCount(ctx context.Context, status *domain.TaskStatus) (int, error) {
	return s.repo.GetTaskCount(ctx, status)
}

const maxBatchSize = 100

func (s *TaskService) BatchCreateTasks(ctx context.Context, cmd *domain.BatchCreateCmd) ([]*domain.Task, error) {
	if len(cmd.Tasks) == 0 {
		return nil, domain.ErrBatchEmpty
	}
	if len(cmd.Tasks) > maxBatchSize {
		return nil, domain.ErrBatchTooLarge
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	now := time.Now()
	tasks := make([]*domain.Task, 0, len(cmd.Tasks))

	for _, c := range cmd.Tasks {
		if c.CronExpr == "" || c.Type == "" || c.Title == "" || c.Payload == "" {
			return nil, domain.ErrValidation
		}
		sch, err := parser.Parse(c.CronExpr)
		if err != nil {
			return nil, domain.ErrInvalidCron
		}

		maxRetries := s.defaultTaskMaxRetries
		if c.MaxRetries != nil {
			maxRetries = *c.MaxRetries
		}

		priority := 5
		if c.Priority != nil {
			priority = *c.Priority
			if priority < 1 {
				priority = 1
			}
			if priority > 10 {
				priority = 10
			}
		}

		t := &domain.Task{
			ID:         uuid.New(),
			Title:      c.Title,
			Type:       c.Type,
			Payload:    []byte(c.Payload),
			CronExpr:   c.CronExpr,
			Status:     domain.TaskStatusPending,
			RetryCount: 0,
			MaxRetries: maxRetries,
			Priority:   priority,
			CreatedAt:  now,
			UpdatedAt:  now,
			NextRunAt:  sch.Next(now),
			ExpiresAt:  c.ExpiresAt,
		}
		tasks = append(tasks, t)
	}

	_, err := s.repo.BatchCreate(ctx, tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) BatchCancelTasks(ctx context.Context, cmd *domain.BatchCancelCmd) (int, error) {
	if len(cmd.IDs) == 0 {
		return 0, domain.ErrBatchEmpty
	}
	if len(cmd.IDs) > maxBatchSize {
		return 0, domain.ErrBatchTooLarge
	}
	return s.repo.BatchCancel(ctx, cmd.IDs)
}

func (s *TaskService) BatchUpdatePriority(ctx context.Context, cmd *domain.BatchUpdatePriorityCmd) (int, error) {
	if len(cmd.IDs) == 0 {
		return 0, domain.ErrBatchEmpty
	}
	if len(cmd.IDs) > maxBatchSize {
		return 0, domain.ErrBatchTooLarge
	}
	if cmd.Priority < 1 || cmd.Priority > 10 {
		return 0, domain.ErrValidation
	}
	return s.repo.BatchUpdatePriority(ctx, cmd.IDs, cmd.Priority)
}

func (s *TaskService) GetAllTasksFiltered(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.GetAllTasksFiltered(ctx, filter)
}

func (s *TaskService) GetTaskCountFiltered(ctx context.Context, filter domain.TaskFilter) (int, error) {
	return s.repo.GetTaskCountFiltered(ctx, filter)
}
