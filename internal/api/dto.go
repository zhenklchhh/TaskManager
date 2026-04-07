package api



import "time"



type CreateTaskRequest struct {

	Title      string     `json:"title" validate:"required"`

	Type       string     `json:"type" validate:"required"`

	Payload    string     `json:"payload" validate:"required"`

	CronExpr   string     `json:"cron_expr" validate:"required"`

	MaxRetries *int       `json:"max_retries,omitempty"`

	Priority   *int       `json:"priority,omitempty"`

	ExpiresAt  *time.Time `json:"expires_at,omitempty"`

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

