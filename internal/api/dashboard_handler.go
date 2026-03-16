package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type DashboardHandler struct {
	taskService *service.TaskService
}

func NewDashboardHandler(taskService *service.TaskService) *DashboardHandler {
	return &DashboardHandler{taskService: taskService}
}

func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.taskService.GetTaskStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *DashboardHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	var status *domain.TaskStatus
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := domain.TaskStatus(statusStr)
		status = &s
	}

	tasks, err := h.taskService.GetAllTasks(r.Context(), limit, offset, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, _ := h.taskService.GetTaskCount(r.Context(), status)

	response := map[string]interface{}{
		"tasks":  tasks,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
