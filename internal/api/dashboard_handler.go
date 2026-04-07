package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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
		"tasks":  toDashboardTaskResponses(tasks),
		"total":  count,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DashboardHandler) GetTasksFiltered(w http.ResponseWriter, r *http.Request) {
	filter := domain.TaskFilter{}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		filter.Limit, _ = strconv.Atoi(limitStr)
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		filter.Offset, _ = strconv.Atoi(offsetStr)
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := domain.TaskStatus(statusStr)
		filter.Status = &s
	}
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		filter.Type = &typeStr
	}
	if pMinStr := r.URL.Query().Get("priority_min"); pMinStr != "" {
		v, _ := strconv.Atoi(pMinStr)
		filter.PriorityMin = &v
	}
	if pMaxStr := r.URL.Query().Get("priority_max"); pMaxStr != "" {
		v, _ := strconv.Atoi(pMaxStr)
		filter.PriorityMax = &v
	}
	if fromStr := r.URL.Query().Get("created_from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			filter.CreatedFrom = &t
		}
	}
	if toStr := r.URL.Query().Get("created_to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			filter.CreatedTo = &t
		}
	}

	tasks, err := h.taskService.GetAllTasksFiltered(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, _ := h.taskService.GetTaskCountFiltered(r.Context(), filter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks":  toDashboardTaskResponses(tasks),
		"total":  count,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}
