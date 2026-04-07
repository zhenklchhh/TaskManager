package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notifService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notifService}
}

type CreateNotificationConfigRequest struct {
	TaskID *string `json:"task_id,omitempty"`
	Type   string  `json:"type" validate:"required"`
	Event  string  `json:"event" validate:"required"`
	Target string  `json:"target" validate:"required"`
}

type NotificationConfigResponse struct {
	ID     string  `json:"id"`
	TaskID *string `json:"task_id,omitempty"`
	Type   string  `json:"type"`
	Event  string  `json:"event"`
	Target string  `json:"target"`
}

func (h *NotificationHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	var req CreateNotificationConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(domain.ErrValidation, w)
		return
	}
	if err := requestValidator.Struct(req); err != nil {
		handleError(domain.ErrValidation, w)
		return
	}

	cfg := &domain.NotificationConfig{
		Type:   domain.NotificationType(req.Type),
		Event:  domain.NotificationEvent(req.Event),
		Target: req.Target,
	}

	if req.TaskID != nil {
		taskID, err := uuid.Parse(*req.TaskID)
		if err != nil {
			handleError(domain.ErrValidation, w)
			return
		}
		cfg.TaskID = &taskID
	}

	if err := h.notificationService.CreateConfig(r.Context(), cfg); err != nil {
		handleError(err, w)
		return
	}

	resp := NotificationConfigResponse{
		ID:     cfg.ID.String(),
		Type:   string(cfg.Type),
		Event:  string(cfg.Event),
		Target: cfg.Target,
	}
	if cfg.TaskID != nil {
		s := cfg.TaskID.String()
		resp.TaskID = &s
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
