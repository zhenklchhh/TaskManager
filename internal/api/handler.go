package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
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
	case errors.Is(err, domain.ErrTaskNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
