package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type Handler struct {
	taskService *service.TaskService
}

func NewHandler(service *service.TaskService) *Handler {
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
	t, err := h.taskService.CreateTask(r.Context(), toCreateTaskCmd(req))
	if err != nil {
		handleError(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(toTaskResponse(t))
	if err != nil {
		handleError(err, w)
		return
	}
}

func (h *Handler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		handleError(errors.New("empty id"), w)
		return
	}
	t, err := h.taskService.GetTaskById(r.Context(), id)
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

func (h *Handler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		handleError(errors.New("empty id"), w)
		return
	}
	var req UpdateTaskInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(errors.New("invalid json"), w)
		return
	}
	err := h.taskService.UpdateTaskStatus(r.Context(), &service.TaskUpdateStatusCmd{
		ID: id,
		Status: req.Status,
	})
	if err != nil {
		handleError(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func handleError(err error, w http.ResponseWriter) {
	switch {
	case errors.Is(err, domain.ErrInvalidCron):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrValidation):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrTaskNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
