package api



import (
	"time"
)



type CreateTaskRequest struct {
	Title      string     `json:"title" validate:"required"`
	Type       string     `json:"type" validate:"required"`
	Payload    string     `json:"payload" validate:"required"`
	CronExpr   string     `json:"cron_expr" validate:"required"`
	MaxRetries *int       `json:"max_retries,omitempty"`
	Priority   *int       `json:"priority,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	GroupID    *string    `json:"group_id,omitempty"`
	AssignedTo *string    `json:"assigned_to,omitempty"`
}



type TaskResponse struct {

	ID        string `json:"id"`

	Title     string `json:"title"`

	Type      string `json:"type"`

	Status    string `json:"status"`

	NextRunAt string `json:"next_run_at"`

}

type DashboardTaskResponse struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	Priority     int        `json:"priority"`
	RetryCount   int        `json:"retry_count"`
	MaxRetries   int        `json:"max_retries"`
	NextRunAt    *time.Time `json:"next_run_at"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	LastErrorMsg string     `json:"last_error_msg,omitempty"`
	CronExpr     string     `json:"cron_expr"`
	Payload      string     `json:"payload"`
	GroupID      *string    `json:"group_id,omitempty"`
	AssignedTo   *string    `json:"assigned_to,omitempty"`
}

type UpdateTaskRequest struct {
	Title      *string `json:"title,omitempty"`
	Payload    *string `json:"payload,omitempty"`
	Priority   *int    `json:"priority,omitempty"`
	CronExpr   *string `json:"cron_expr,omitempty"`
	MaxRetries *int    `json:"max_retries,omitempty"`
	GroupID    *string `json:"group_id"`
	AssignedTo *string `json:"assigned_to,omitempty"`
}

type TaskExecutionResponse struct {
	ID         string     `json:"id"`
	TaskID     string     `json:"task_id"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Status     string     `json:"status"`
	Error      string     `json:"error,omitempty"`
	Output     string     `json:"output,omitempty"`
	WorkerID   string     `json:"worker_id,omitempty"`
	DurationMs int64      `json:"duration_ms"`
}

type BatchCreateRequest struct {
	Tasks []CreateTaskRequest `json:"tasks" validate:"required,dive"`
}

type BatchCancelRequest struct {
	IDs []string `json:"ids" validate:"required"`
}

type BatchUpdatePriorityRequest struct {
	IDs      []string `json:"ids" validate:"required"`
	Priority int      `json:"priority" validate:"required,min=1,max=10"`
}

type BatchResponse struct {
	Affected int    `json:"affected"`
	Message  string `json:"message"`
}

