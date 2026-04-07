package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type DependencyHandler struct {
	dependencyService *service.DependencyService
}

func NewDependencyHandler(depService *service.DependencyService) *DependencyHandler {
	return &DependencyHandler{dependencyService: depService}
}

type AddDependencyRequest struct {
	TaskID      string `json:"task_id" validate:"required"`
	DependsOnID string `json:"depends_on_id" validate:"required"`
	Condition   string `json:"condition,omitempty"`
}

type DependencyResponse struct {
	ID          string `json:"id"`
	TaskID      string `json:"task_id"`
	DependsOnID string `json:"depends_on_id"`
	Condition   string `json:"condition"`
}

func (h *DependencyHandler) AddDependency(w http.ResponseWriter, r *http.Request) {
	var req AddDependencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(domain.ErrValidation, w)
		return
	}

	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	dependsOnID, err := uuid.Parse(req.DependsOnID)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}

	condition := domain.DependencyCondition(req.Condition)
	if condition == "" {
		condition = domain.ConditionCompleted
	}

	dep, err := h.dependencyService.AddDependency(r.Context(), taskID, dependsOnID, condition)
	if err != nil {
		handleError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(DependencyResponse{
		ID:          dep.ID.String(),
		TaskID:      dep.TaskID.String(),
		DependsOnID: dep.DependsOnID.String(),
		Condition:   string(dep.Condition),
	})
}

func (h *DependencyHandler) GetDependencies(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}

	deps, err := h.dependencyService.GetDependencies(r.Context(), taskID)
	if err != nil {
		handleError(err, w)
		return
	}

	responses := make([]DependencyResponse, 0, len(deps))
	for _, d := range deps {
		responses = append(responses, DependencyResponse{
			ID:          d.ID.String(),
			TaskID:      d.TaskID.String(),
			DependsOnID: d.DependsOnID.String(),
			Condition:   string(d.Condition),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (h *DependencyHandler) GetDependents(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}

	deps, err := h.dependencyService.GetDependents(r.Context(), taskID)
	if err != nil {
		handleError(err, w)
		return
	}

	responses := make([]DependencyResponse, 0, len(deps))
	for _, d := range deps {
		responses = append(responses, DependencyResponse{
			ID:          d.ID.String(),
			TaskID:      d.TaskID.String(),
			DependsOnID: d.DependsOnID.String(),
			Condition:   string(d.Condition),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (h *DependencyHandler) RemoveDependency(w http.ResponseWriter, r *http.Request) {
	depIDStr := chi.URLParam(r, "dep_id")
	depID, err := uuid.Parse(depIDStr)
	if err != nil {
		handleError(domain.ErrValidation, w)
		return
	}

	if err := h.dependencyService.RemoveDependency(r.Context(), depID); err != nil {
		handleError(err, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DependencyHandler) GetChildTasks(w http.ResponseWriter, r *http.Request) {
	parentIDStr := chi.URLParam(r, "id")
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		handleError(errors.New("invalid parent id"), w)
		return
	}

	children, err := h.dependencyService.GetChildTasks(r.Context(), parentID)
	if err != nil {
		handleError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDashboardTaskResponses(children))
}
