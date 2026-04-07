package api

import (
	"context"
	"errors"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type stubTaskService struct {
	createTaskFn  func(ctx context.Context, cmd *domain.TaskCreateCmd) (*domain.Task, error)
	getTaskByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Task, error)
}

func (s stubTaskService) CreateTask(ctx context.Context, cmd *domain.TaskCreateCmd) (*domain.Task, error) {
	if s.createTaskFn != nil {
		return s.createTaskFn(ctx, cmd)
	}
	t := &domain.Task{
		ID:    uuid.New(),
		Title: cmd.Title,
	}
	return t, nil
}

func (s stubTaskService) GetTaskById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	if s.getTaskByIDFn != nil {
		return s.getTaskByIDFn(ctx, id)
	}
	return nil, errors.New("GetTaskById not stubbed")
}

func TestCreateTask(t *testing.T) {
	type testCase struct {
		name           string
		body           string
		serviceErr     error
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name: "valid request",
			body: `{"title":"test task","type":"email","payload":"{\"to\":\"user@example.com\"}","cron_expr":"*/5 * * * *"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "validation error - empty title",
			body: `{"title":"","type":"email","payload":"{\"to\":\"user@example.com\"}","cron_expr":"*/5 * * * *"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid parameters\n",
		},
		{
			name: "validation error - invalid cron",
			body: `{"title":"test task","type":"email","payload":"{\"to\":\"user@example.com\"}","cron_expr":"not-a-cron"}`,
			serviceErr:     domain.ErrInvalidCron,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid cron expression\n",
		},
		{
			name: "service error - repository failure",
			body:       `{"title":"test task","type":"email","payload":"{\"to\":\"user@example.com\"}","cron_expr":"*/5 * * * *"}`,
			serviceErr: errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal error\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := stubTaskService{
				createTaskFn: func(ctx context.Context, cmd *domain.TaskCreateCmd) (*domain.Task, error) {
					if tt.serviceErr != nil {
						return nil, tt.serviceErr
					}
					return &domain.Task{
						ID:    uuid.New(),
						Title: cmd.Title,
					}, nil
				},
			}

			handler := &Handler{taskService: svc}
			router := Routes(handler, &HealthChecker{}, &DashboardHandler{}, &BatchHandler{}, &DependencyHandler{}, &NotificationHandler{})

			var bodyReader *bytes.Reader
			if tt.body != "" {
				bodyReader = bytes.NewReader([]byte(tt.body))
			} else {
				bodyReader = bytes.NewReader(nil)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bodyReader)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" {
				if rr.Body.String() != tt.expectedBody {
					t.Fatalf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
				}
			} else if tt.expectedStatus == http.StatusCreated {
				var resp struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("expected valid JSON response, got error: %v, body: %s", err, rr.Body.String())
				}
				if resp.ID == "" {
					t.Fatalf("expected non-empty id in response")
				}
			}
		})
	}
}

func TestGetTaskById(t *testing.T) {
	type testCase struct {
		name           string
		id             string
		serviceErr     error
		task           *domain.Task
		expectedStatus int
		expectedBody   string
	}

	validID := uuid.New()

	tests := []testCase{
		{
			name:           "valid id - task found",
			id:             validID.String(),
			task:           &domain.Task{ID: validID, Title: "test task", Status: domain.TaskStatusPending},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid format",
			id:             "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid parameters\n",
		},
		{
			name:           "task not found",
			id:             validID.String(),
			serviceErr:     domain.ErrTaskNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "task not found\n",
		},
		{
			name:           "service internal error",
			id:             validID.String(),
			serviceErr:     errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal error\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := stubTaskService{
				getTaskByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
					if tt.serviceErr != nil {
						return nil, tt.serviceErr
					}
					if tt.task != nil {
						return tt.task, nil
					}
					return &domain.Task{ID: id, Title: "default"}, nil
				},
			}

			handler := &Handler{taskService: svc}
			router := Routes(handler, &HealthChecker{}, &DashboardHandler{}, &BatchHandler{}, &DependencyHandler{}, &NotificationHandler{})

			path := "/api/v1/tasks/" + tt.id
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" {
				if rr.Body.String() != tt.expectedBody {
					t.Fatalf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
				}
			} else if tt.expectedStatus == http.StatusOK {
				var resp TaskResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("expected valid JSON response, got error: %v, body: %s", err, rr.Body.String())
				}
				if resp.ID != tt.id {
					t.Fatalf("expected id %q, got %q", tt.id, resp.ID)
				}
			}
		})
	}
}