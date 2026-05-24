package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/metrics"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type Handler struct {
	taskService service.TaskServiceInterface
}

var requestValidator = validator.New()

func NewHandler(service service.TaskServiceInterface) *Handler {
	return &Handler{
		taskService: service,
	}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(errors.New("invalid json"), w)
		return
	}
	if err := requestValidator.Struct(req); err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	cmd := toCreateTaskCmd(req)
	// Inject company and user from auth context
	companyID := GetCompanyID(r.Context())
	if companyID != nil {
		cmd.CompanyID = companyID
	}
	userID := GetUserID(r.Context())
	if userID != (uuid.UUID{}) {
		cmd.CreatedBy = &userID
	}
	t, err := h.taskService.CreateTask(r.Context(), cmd)
	if err != nil {
		handleError(err, w)
		return
	}
	metrics.RecordTaskCreated(t.Type, string(t.Status))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(toTaskResponse(t))
	if err != nil {
		handleError(err, w)
		return
	}
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	stringID := chi.URLParam(r, "id")
	if stringID == "" {
		handleError(errors.New("empty id"), w)
		return
	}
	taskID, err := uuid.Parse(stringID)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(errors.New("invalid json"), w)
		return
	}
	cmd := toUpdateTaskCmd(taskID, req)
	t, err := h.taskService.UpdateTask(r.Context(), cmd)
	if err != nil {
		handleError(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toDashboardTaskResponse(t))
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	stringID := chi.URLParam(r, "id")
	if stringID == "" {
		handleError(errors.New("empty id"), w)
		return
	}
	taskID, err := uuid.Parse(stringID)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	if err := h.taskService.DeleteTask(r.Context(), taskID); err != nil {
		handleError(err, w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CancelTask(w http.ResponseWriter, r *http.Request) {
	stringID := chi.URLParam(r, "id")
	if stringID == "" {
		handleError(errors.New("empty id"), w)
		return
	}
	taskID, err := uuid.Parse(stringID)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	if err := h.taskService.CancelTask(r.Context(), taskID); err != nil {
		handleError(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

func (h *Handler) RetryTask(w http.ResponseWriter, r *http.Request) {
	stringID := chi.URLParam(r, "id")
	if stringID == "" {
		handleError(errors.New("empty id"), w)
		return
	}
	taskID, err := uuid.Parse(stringID)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	if err := h.taskService.ManualRetry(r.Context(), taskID); err != nil {
		handleError(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
}

func (h *Handler) GetTaskLogs(w http.ResponseWriter, r *http.Request) {
	stringID := chi.URLParam(r, "id")
	if stringID == "" {
		handleError(errors.New("empty id"), w)
		return
	}
	taskID, err := uuid.Parse(stringID)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	executions, err := h.taskService.GetTaskExecutions(r.Context(), taskID)
	if err != nil {
		handleError(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"executions": toExecutionResponses(executions),
		"total":      len(executions),
	})
}

func (h *Handler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	stringID := chi.URLParam(r, "id")
	if stringID == "" {
		handleError(errors.New("empty id"), w)
		return
	}
	taskID, err := uuid.Parse(stringID)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	t, err := h.taskService.GetTaskById(r.Context(), taskID)
	if err != nil {
		handleError(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(toTaskResponse(t))
	if err != nil {
		handleError(err, w)
	}
}

func handleError(err error, w http.ResponseWriter) {
	switch {
	case errors.Is(err, domain.ErrInvalidCron):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrValidation):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrBatchEmpty):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrBatchTooLarge):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrTaskNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrTaskDeleted):
		http.Error(w, err.Error(), http.StatusGone)
	case errors.Is(err, domain.ErrTaskNotCancellable):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrMaxRetriesExceeded):
		http.Error(w, err.Error(), http.StatusGone)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
