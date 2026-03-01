package api

import "time"

type CreateTaskRequest struct {
	Title      string     `json:"title" validate:"required"`
	Type       string     `json:"type" validate:"required"`
	Payload    string     `json:"payload" validate:"required"`
	CronExpr   string     `json:"cron_expr" validate:"required"`
	MaxRetries *int       `json:"max_retries,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type TaskResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	NextRunAt string `json:"next_run_at"`
}
