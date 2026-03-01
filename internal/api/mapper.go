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
		ExpiresAt:  t.ExpiresAt,
	}
}

func toTaskResponse(t *domain.Task) *TaskResponse {
	return &TaskResponse{
		t.ID.String(),
		t.Title,
		string(t.Status),
		t.NextRunAt.String(),
		t.Type,
	}
}
