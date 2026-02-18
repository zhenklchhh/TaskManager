package http

import (
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type CreateTaskRequest struct {
	Title    string `json:"title" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Payload  string `json:"payload" validate:"required"`
	CronExpr string `json:"cron_expr" validate:"required, cron"`
}

type TaskResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	NextRunAt string `json:"next_run_at"`
}

type UpdateTaskInfo struct {
	Status domain.TaskStatus `json:"status"`
}
