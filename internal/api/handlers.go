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

// todo: use router to get id from url path
// implement helper function to handle errors to prevent duplicate of code
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	t, err := h.taskService.CreateTask(r.Context(), toCreateTaskCmd(req))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCron):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, domain.ErrValidation):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toTaskResponse(t))
}

func (h *Handler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "internal error: invalid id", http.StatusBadRequest)
	}
	t, err := h.taskService.GetTaskById(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTaskNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(toTaskResponse(t))
	if err != nil {

	}
}
