package api

import (
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

func toCreateTaskCmd(t CreateTaskRequest) *domain.TaskCreateCmd {
	return &domain.TaskCreateCmd{
		Title:      t.Title,
		Type:       t.Type,
		Payload:    t.Payload,
		CronExpr:   t.CronExpr,
		MaxRetries: t.MaxRetries,
		Priority:   t.Priority,
		ExpiresAt:  t.ExpiresAt,
	}
}

func toTaskResponse(t *domain.Task) *TaskResponse {
	return &TaskResponse{
		t.ID.String(),
		t.Title,
		t.Type,
		string(t.Status),
		t.NextRunAt.String(),
	}
}

func toDashboardTaskResponse(t *domain.Task) *DashboardTaskResponse {
	nextRunAt := t.NextRunAt
	createdAt := t.CreatedAt
	updatedAt := t.UpdatedAt
	return &DashboardTaskResponse{
		ID:           t.ID.String(),
		Title:        t.Title,
		Type:         t.Type,
		Status:       string(t.Status),
		Priority:     t.Priority,
		RetryCount:   t.RetryCount,
		MaxRetries:   t.MaxRetries,
		NextRunAt:    &nextRunAt,
		CreatedAt:    &createdAt,
		UpdatedAt:    &updatedAt,
		LastErrorMsg: t.LastErrorMsg,
		CronExpr:     t.CronExpr,
	}
}

func toDashboardTaskResponses(tasks []*domain.Task) []*DashboardTaskResponse {
	result := make([]*DashboardTaskResponse, len(tasks))
	for i, t := range tasks {
		result[i] = toDashboardTaskResponse(t)
	}
	return result
}
