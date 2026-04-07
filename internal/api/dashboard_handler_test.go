package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type stubDashboardRepo struct {
	stats     *domain.TaskStats
	statsErr  error
	tasks     []*domain.Task
	tasksErr  error
	count     int
	countErr  error
}

func (s *stubDashboardRepo) Create(ctx context.Context, t *domain.Task) error { return nil }
func (s *stubDashboardRepo) GetTaskById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return nil, nil
}
func (s *stubDashboardRepo) GetPendingTasks(ctx context.Context, limit int) ([]uuid.UUID, error) {
	return nil, nil
}
func (s *stubDashboardRepo) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error {
	return nil
}
func (s *stubDashboardRepo) UpdateTaskForRetry(ctx context.Context, id uuid.UUID, lastErrorMsg string, status domain.TaskStatus, retries int, nextRunAt time.Time) error {
	return nil
}
func (s *stubDashboardRepo) UpdateStaleTasksToPending(ctx context.Context, threshold time.Duration) (int, error) {
	return 0, nil
}

func (s *stubDashboardRepo) GetTaskStats(ctx context.Context) (*domain.TaskStats, error) {
	if s.statsErr != nil {
		return nil, s.statsErr
	}
	return s.stats, nil
}

func (s *stubDashboardRepo) GetAllTasks(ctx context.Context, limit, offset int, status *domain.TaskStatus) ([]*domain.Task, error) {
	if s.tasksErr != nil {
		return nil, s.tasksErr
	}
	return s.tasks, nil
}

func (s *stubDashboardRepo) GetTaskCount(ctx context.Context, status *domain.TaskStatus) (int, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
	return s.count, nil
}

func (s *stubDashboardRepo) BatchCreate(ctx context.Context, tasks []*domain.Task) (int, error) {
	return len(tasks), nil
}
func (s *stubDashboardRepo) BatchCancel(ctx context.Context, ids []uuid.UUID) (int, error) {
	return len(ids), nil
}
func (s *stubDashboardRepo) BatchUpdatePriority(ctx context.Context, ids []uuid.UUID, priority int) (int, error) {
	return len(ids), nil
}
func (s *stubDashboardRepo) GetAllTasksFiltered(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, error) {
	return []*domain.Task{}, nil
}
func (s *stubDashboardRepo) GetTaskCountFiltered(ctx context.Context, filter domain.TaskFilter) (int, error) {
	return 0, nil
}

func TestGetStats_EmptyDB(t *testing.T) {
	repo := &stubDashboardRepo{
		stats: &domain.TaskStats{
			TotalTasks:      0,
			PendingTasks:    0,
			ScheduledTasks:  0,
			RunningTasks:    0,
			CompletedTasks:  0,
			FailedTasks:     0,
			TasksByType:     map[string]int{},
			TasksByPriority: map[int]int{},
		},
	}
	svc := service.NewTaskService(repo, 3)
	dh := NewDashboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
	rr := httptest.NewRecorder()

	dh.GetStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var stats domain.TaskStats
	if err := json.Unmarshal(rr.Body.Bytes(), &stats); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if stats.TotalTasks != 0 {
		t.Fatalf("expected 0 total tasks, got %d", stats.TotalTasks)
	}
}

func TestGetStats_WithTasks(t *testing.T) {
	repo := &stubDashboardRepo{
		stats: &domain.TaskStats{
			TotalTasks:      5,
			PendingTasks:    2,
			ScheduledTasks:  1,
			RunningTasks:    1,
			CompletedTasks:  1,
			FailedTasks:     0,
			TasksByType:     map[string]int{"email": 3, "webhook": 2},
			TasksByPriority: map[int]int{1: 2, 5: 3},
		},
	}
	svc := service.NewTaskService(repo, 3)
	dh := NewDashboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
	rr := httptest.NewRecorder()

	dh.GetStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var stats domain.TaskStats
	if err := json.Unmarshal(rr.Body.Bytes(), &stats); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if stats.TotalTasks != 5 {
		t.Fatalf("expected 5 total tasks, got %d", stats.TotalTasks)
	}
	if stats.PendingTasks != 2 {
		t.Fatalf("expected 2 pending tasks, got %d", stats.PendingTasks)
	}
}

func TestGetStats_Error(t *testing.T) {
	repo := &stubDashboardRepo{
		statsErr: errors.New("db error"),
	}
	svc := service.NewTaskService(repo, 3)
	dh := NewDashboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
	rr := httptest.NewRecorder()

	dh.GetStats(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestGetTasks_DefaultPagination(t *testing.T) {
	tasks := []*domain.Task{
		{ID: uuid.New(), Title: "Task 1", Status: domain.TaskStatusPending},
		{ID: uuid.New(), Title: "Task 2", Status: domain.TaskStatusCompleted},
	}
	repo := &stubDashboardRepo{
		tasks: tasks,
		count: 2,
	}
	svc := service.NewTaskService(repo, 3)
	dh := NewDashboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/tasks", nil)
	rr := httptest.NewRecorder()

	dh.GetTasks(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["total"].(float64) != 2 {
		t.Fatalf("expected total=2, got %v", resp["total"])
	}
	if resp["limit"].(float64) != 20 {
		t.Fatalf("expected default limit=20, got %v", resp["limit"])
	}
	if resp["offset"].(float64) != 0 {
		t.Fatalf("expected default offset=0, got %v", resp["offset"])
	}
}

func TestGetTasks_WithPagination(t *testing.T) {
	repo := &stubDashboardRepo{
		tasks: []*domain.Task{},
		count: 50,
	}
	svc := service.NewTaskService(repo, 3)
	dh := NewDashboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/tasks?limit=10&offset=20", nil)
	rr := httptest.NewRecorder()

	dh.GetTasks(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["limit"].(float64) != 10 {
		t.Fatalf("expected limit=10, got %v", resp["limit"])
	}
	if resp["offset"].(float64) != 20 {
		t.Fatalf("expected offset=20, got %v", resp["offset"])
	}
}

func TestGetTasks_WithStatusFilter(t *testing.T) {
	tasks := []*domain.Task{
		{ID: uuid.New(), Title: "Pending Task", Status: domain.TaskStatusPending},
	}
	repo := &stubDashboardRepo{
		tasks: tasks,
		count: 1,
	}
	svc := service.NewTaskService(repo, 3)
	dh := NewDashboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/tasks?status=pending", nil)
	rr := httptest.NewRecorder()

	dh.GetTasks(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["total"].(float64) != 1 {
		t.Fatalf("expected total=1, got %v", resp["total"])
	}
}

func TestGetTasks_Error(t *testing.T) {
	repo := &stubDashboardRepo{
		tasksErr: errors.New("db error"),
	}
	svc := service.NewTaskService(repo, 3)
	dh := NewDashboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/tasks", nil)
	rr := httptest.NewRecorder()

	dh.GetTasks(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
