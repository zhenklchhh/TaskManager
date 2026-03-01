package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type stubRepository struct {
	createFn func(ctx context.Context, t *domain.Task) error

	getTaskByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Task, error)

	getPendingTasksFn func(ctx context.Context, limit int) ([]uuid.UUID, error)

	updateTaskStatusFn func(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error

	updateTaskForRetryFn func(ctx context.Context, id uuid.UUID, lastErrorMsg string, status domain.TaskStatus, retries int, nextRunAt time.Time) error

	updateStaleTasksToPendingFn func(ctx context.Context, threshold time.Duration) (int, error)

	createCalled int
	created      *domain.Task

	updateStatusCalled int
	updatedStatusID    uuid.UUID
	updatedStatus      domain.TaskStatus

	updateForRetryCalled       int
	updateForRetryID           uuid.UUID
	updateForRetryStatus       domain.TaskStatus
	updateForRetryRetries      int
	updateForRetryNextRunAt    time.Time
	updateForRetryLastErrorMsg string
}

func (s *stubRepository) Create(ctx context.Context, t *domain.Task) error {
	s.createCalled++
	s.created = t
	if s.createFn != nil {
		return s.createFn(ctx, t)
	}
	return nil
}

func (s *stubRepository) GetTaskById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	if s.getTaskByIDFn != nil {
		return s.getTaskByIDFn(ctx, id)
	}
	return nil, errors.New("GetTaskById not stubbed")
}

func (s *stubRepository) GetPendingTasks(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if s.getPendingTasksFn != nil {
		return s.getPendingTasksFn(ctx, limit)
	}
	return nil, errors.New("GetPendingTasks not stubbed")
}

func (s *stubRepository) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error {
	s.updateStatusCalled++
	s.updatedStatusID = id
	s.updatedStatus = status
	if s.updateTaskStatusFn != nil {
		return s.updateTaskStatusFn(ctx, id, status)
	}
	return nil
}

func (s *stubRepository) UpdateTaskForRetry(ctx context.Context, id uuid.UUID, lastErrorMsg string, status domain.TaskStatus, retries int, nextRunAt time.Time) error {
	s.updateForRetryCalled++
	s.updateForRetryID = id
	s.updateForRetryLastErrorMsg = lastErrorMsg
	s.updateForRetryStatus = status
	s.updateForRetryRetries = retries
	s.updateForRetryNextRunAt = nextRunAt
	if s.updateTaskForRetryFn != nil {
		return s.updateTaskForRetryFn(ctx, id, lastErrorMsg, status, retries, nextRunAt)
	}
	return nil
}

func (s *stubRepository) UpdateStaleTasksToPending (ctx context.Context, threshold time.Duration) (int, error) {
	if s.updateStaleTasksToPendingFn != nil {
		return s.updateStaleTasksToPendingFn(ctx, threshold)
	}
	return 0, nil
}

const defaultTaskMaxRetries = 3

func TestTaskService_CreateTask(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		cmd  *domain.TaskCreateCmd

		repoErr          error
		wantErr          error
		wantCreateCalled bool
	}

	tests := []testCase{
		{
			name: "validation error: empty fields",
			cmd: &domain.TaskCreateCmd{
				Title:    "",
				Type:     "test",
				Payload:  "{}",
				CronExpr: "*/1 * * * *",
			},
			wantErr:          domain.ErrValidation,
			wantCreateCalled: false,
		},
		{
			name: "invalid cron",
			cmd: &domain.TaskCreateCmd{
				Title:    "t",
				Type:     "test",
				Payload:  "{}",
				CronExpr: "not a cron",
			},
			wantErr:          domain.ErrInvalidCron,
			wantCreateCalled: false,
		},
		{
			name: "repo error is propagated",
			cmd: &domain.TaskCreateCmd{
				Title:    "t",
				Type:     "test",
				Payload:  "{}",
				CronExpr: "*/1 * * * *",
			},
			repoErr:          errors.New("db down"),
			wantErr:          errors.New("db down"),
			wantCreateCalled: true,
		},
		{
			name: "success creates task with defaults",
			cmd: &domain.TaskCreateCmd{
				Title:    "hello",
				Type:     "test",
				Payload:  "{}",
				CronExpr: "*/1 * * * *",
			},
			wantErr:          nil,
			wantCreateCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubRepository{
				createFn: func(ctx context.Context, task *domain.Task) error {
					return tt.repoErr
				},
			}
			svc := NewTaskService(repo, defaultTaskMaxRetries)

			now := time.Now()
			task, err := svc.CreateTask(context.Background(), tt.cmd)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if tt.wantCreateCalled && repo.createCalled != 1 {
				t.Fatalf("expected Create to be called once, got %d", repo.createCalled)
			}
			if !tt.wantCreateCalled && repo.createCalled != 0 {
				t.Fatalf("expected Create not to be called, got %d", repo.createCalled)
			}

			if tt.wantErr != nil {
				if task != nil {
					t.Fatalf("expected nil task on error, got %+v", task)
				}
				return
			}

			if task == nil {
				t.Fatalf("expected non-nil task")
			}
			if repo.created == nil {
				t.Fatalf("expected repo to receive created task")
			}
			if task.ID == uuid.Nil {
				t.Fatalf("expected non-nil UUID")
			}
			if task.Status != domain.TaskStatusPending {
				t.Fatalf("expected status %q, got %q", domain.TaskStatusPending, task.Status)
			}
			if task.RetryCount != 0 {
				t.Fatalf("expected RetryCount=0, got %d", task.RetryCount)
			}
			if task.MaxRetries != 3 {
				t.Fatalf("expected MaxRetries=3, got %d", task.MaxRetries)
			}
			if task.Title != tt.cmd.Title || task.Type != tt.cmd.Type || string(task.Payload) != tt.cmd.Payload || task.CronExpr != tt.cmd.CronExpr {
				t.Fatalf("created task does not match cmd: %+v", task)
			}
			if task.NextRunAt.IsZero() {
				t.Fatalf("expected NextRunAt to be set")
			}
			if !task.NextRunAt.After(now) {
				t.Fatalf("expected NextRunAt after now; now=%v next=%v", now, task.NextRunAt)
			}
		})
	}
}

func TestTaskService_ProcessPendingTasks(t *testing.T) {
	t.Parallel()

	want := []uuid.UUID{uuid.New(), uuid.New()}
	repo := &stubRepository{
		getPendingTasksFn: func(ctx context.Context, limit int) ([]uuid.UUID, error) {
			// in future limit will be value from env file
			if limit != 50 {
				t.Fatalf("expected limit=50, got %d", limit)
			}
			return want, nil
		},
	}
	svc := NewTaskService(repo, defaultTaskMaxRetries)

	got, err := svc.ProcessPendingTasks(context.Background(), 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d ids, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected id[%d]=%v, got %v", i, want[i], got[i])
		}
	}
}

func TestTaskService_RetryTask(t *testing.T) {
	t.Parallel()

	t.Run("propagates GetTaskById error", func(t *testing.T) {
		repo := &stubRepository{
			getTaskByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
				return nil, domain.ErrTaskNotFound
			},
		}
		svc := NewTaskService(repo, defaultTaskMaxRetries)

		err := svc.RetryTask(context.Background(), uuid.New(), errors.New("boom"))
		if !errors.Is(err, domain.ErrTaskNotFound) {
			t.Fatalf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("when retries exceeded -> sets failed", func(t *testing.T) {
		id := uuid.New()
		repo := &stubRepository{
			getTaskByIDFn: func(ctx context.Context, gotID uuid.UUID) (*domain.Task, error) {
				if gotID != id {
					t.Fatalf("unexpected id: %v", gotID)
				}
				return &domain.Task{
					ID:         id,
					RetryCount: 3,
					MaxRetries: 3,
				}, nil
			},
		}
		svc := NewTaskService(repo, defaultTaskMaxRetries)

		err := svc.RetryTask(context.Background(), id, errors.New("boom"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.updateStatusCalled != 1 {
			t.Fatalf("expected UpdateTaskStatus to be called once, got %d", repo.updateStatusCalled)
		}
		if repo.updatedStatusID != id {
			t.Fatalf("expected updated id %v, got %v", id, repo.updatedStatusID)
		}
		if repo.updatedStatus != domain.TaskStatusFailed {
			t.Fatalf("expected status %q, got %q", domain.TaskStatusFailed, repo.updatedStatus)
		}
		if repo.updateForRetryCalled != 0 {
			t.Fatalf("expected UpdateTaskForRetry not to be called, got %d", repo.updateForRetryCalled)
		}
	})

	t.Run("schedules retry with exponential backoff and stores error", func(t *testing.T) {
		id := uuid.New()
		taskErr := errors.New("upstream timeout")

		repo := &stubRepository{
			getTaskByIDFn: func(ctx context.Context, gotID uuid.UUID) (*domain.Task, error) {
				return &domain.Task{
					ID:         id,
					RetryCount: 0,
					MaxRetries: 3,
				}, nil
			},
		}
		svc := NewTaskService(repo, defaultTaskMaxRetries)

		start := time.Now().UTC()
		err := svc.RetryTask(context.Background(), id, taskErr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if repo.updateForRetryCalled != 1 {
			t.Fatalf("expected UpdateTaskForRetry to be called once, got %d", repo.updateForRetryCalled)
		}
		if repo.updateForRetryID != id {
			t.Fatalf("expected id=%v, got %v", id, repo.updateForRetryID)
		}
		if repo.updateForRetryStatus != domain.TaskStatusPending {
			t.Fatalf("expected status %q, got %q", domain.TaskStatusPending, repo.updateForRetryStatus)
		}
		if repo.updateForRetryRetries != 1 {
			t.Fatalf("expected retries=1, got %d", repo.updateForRetryRetries)
		}
		if repo.updateForRetryLastErrorMsg != taskErr.Error() {
			t.Fatalf("expected last error %q, got %q", taskErr.Error(), repo.updateForRetryLastErrorMsg)
		}

		wantMin := start.Add(time.Minute)
		wantMax := start.Add(2 * time.Minute)
		if repo.updateForRetryNextRunAt.Before(wantMin) || repo.updateForRetryNextRunAt.After(wantMax) {
			t.Fatalf("unexpected next_run_at: got=%v expected between %v and %v", repo.updateForRetryNextRunAt, wantMin, wantMax)
		}
	})

}
