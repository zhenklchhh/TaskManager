package api

import (
	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

func toCreateTaskCmd(t CreateTaskRequest) *domain.TaskCreateCmd {
	cmd := &domain.TaskCreateCmd{
		Title:      t.Title,
		Type:       t.Type,
		Payload:    t.Payload,
		CronExpr:   t.CronExpr,
		MaxRetries: t.MaxRetries,
		Priority:   t.Priority,
		ExpiresAt:  t.ExpiresAt,
	}
	if t.GroupID != nil {
		if id, err := uuid.Parse(*t.GroupID); err == nil {
			cmd.GroupID = &id
		}
	}
	if t.AssignedTo != nil {
		if id, err := uuid.Parse(*t.AssignedTo); err == nil {
			cmd.AssignedTo = &id
		}
	}
	return cmd
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
	resp := &DashboardTaskResponse{
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
		Payload:      string(t.Payload),
	}
	if t.GroupID != nil {
		s := t.GroupID.String()
		resp.GroupID = &s
	}
	if t.AssignedTo != nil {
		s := t.AssignedTo.String()
		resp.AssignedTo = &s
	}
	return resp
}

func toUpdateTaskCmd(id uuid.UUID, req UpdateTaskRequest) *domain.TaskUpdateCmd {
	cmd := &domain.TaskUpdateCmd{
		ID:         id,
		Title:      req.Title,
		Payload:    req.Payload,
		Priority:   req.Priority,
		CronExpr:   req.CronExpr,
		MaxRetries: req.MaxRetries,
	}
	if req.GroupID != nil {
		if *req.GroupID == "" {
			cmd.ClearGroup = true
		} else if gid, err := uuid.Parse(*req.GroupID); err == nil {
			cmd.GroupID = &gid
		}
	}
	if req.AssignedTo != nil {
		if aid, err := uuid.Parse(*req.AssignedTo); err == nil {
			cmd.AssignedTo = &aid
		}
	}
	return cmd
}

func toExecutionResponse(r *domain.TaskRun) *TaskExecutionResponse {
	return &TaskExecutionResponse{
		ID:         r.ID.String(),
		TaskID:     r.TaskID.String(),
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
		Status:     string(r.Status),
		Error:      r.Error,
		Output:     r.Output,
		WorkerID:   r.WorkerID,
		DurationMs: r.DurationMs,
	}
}

func toExecutionResponses(runs []*domain.TaskRun) []*TaskExecutionResponse {
	result := make([]*TaskExecutionResponse, len(runs))
	for i, r := range runs {
		result[i] = toExecutionResponse(r)
	}
	return result
}

func toDashboardTaskResponses(tasks []*domain.Task) []*DashboardTaskResponse {
	result := make([]*DashboardTaskResponse, len(tasks))
	for i, t := range tasks {
		result[i] = toDashboardTaskResponse(t)
	}
	return result
}
