package api

import (
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

func toCreateTaskCmd(t CreateTaskRequest) *service.TaskCreateCmd {
	return &service.TaskCreateCmd{
		Title: t.Title,
		Type: t.Type,
		Payload: t.Payload,
		CronExpr: t.CronExpr,
	}
}

func toTaskResponse(t *domain.Task) *TaskResponse {
	return &TaskResponse{
		t.ID.String(),
		t.Title,
		string(t.Status),
		t.NextRunAt.String(),
	}
}