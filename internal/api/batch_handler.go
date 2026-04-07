package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type BatchHandler struct {
	taskService *service.TaskService
}

func NewBatchHandler(taskService *service.TaskService) *BatchHandler {
	return &BatchHandler{taskService: taskService}
}

func (h *BatchHandler) BatchCreate(w http.ResponseWriter, r *http.Request) {
	var req BatchCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	if err := requestValidator.Struct(req); err != nil {
		handleError(domain.ErrValidation, w)
		return
	}

	cmds := make([]domain.TaskCreateCmd, 0, len(req.Tasks))
	for _, t := range req.Tasks {
		cmds = append(cmds, domain.TaskCreateCmd{
			Title:      t.Title,
			Type:       t.Type,
			Payload:    t.Payload,
			CronExpr:   t.CronExpr,
			MaxRetries: t.MaxRetries,
			Priority:   t.Priority,
			ExpiresAt:  t.ExpiresAt,
		})
	}

	tasks, err := h.taskService.BatchCreateTasks(r.Context(), &domain.BatchCreateCmd{Tasks: cmds})
	if err != nil {
		handleError(err, w)
		return
	}

	responses := make([]*TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		responses = append(responses, toTaskResponse(t))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"created": len(responses),
		"tasks":   responses,
	})
}

func (h *BatchHandler) BatchCancel(w http.ResponseWriter, r *http.Request) {
	var req BatchCancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(domain.ErrValidation, w)
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, s := range req.IDs {
		id, err := uuid.Parse(s)
		if err != nil {
			handleError(domain.ErrValidation, w)
			return
		}
		ids = append(ids, id)
	}

	affected, err := h.taskService.BatchCancelTasks(r.Context(), &domain.BatchCancelCmd{IDs: ids})
	if err != nil {
		handleError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BatchResponse{
		Affected: affected,
		Message:  fmt.Sprintf("cancelled %d tasks", affected),
	})
}

func (h *BatchHandler) BatchUpdatePriority(w http.ResponseWriter, r *http.Request) {
	var req BatchUpdatePriorityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(domain.ErrValidation, w)
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, s := range req.IDs {
		id, err := uuid.Parse(s)
		if err != nil {
			handleError(domain.ErrValidation, w)
			return
		}
		ids = append(ids, id)
	}

	affected, err := h.taskService.BatchUpdatePriority(r.Context(), &domain.BatchUpdatePriorityCmd{
		IDs:      ids,
		Priority: req.Priority,
	})
	if err != nil {
		handleError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BatchResponse{
		Affected: affected,
		Message:  fmt.Sprintf("updated priority for %d tasks", affected),
	})
}
