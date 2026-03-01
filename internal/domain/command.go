package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskCreateCmd struct {
	Title      string
	Type       string
	Payload    string
	CronExpr   string
	MaxRetries *int
	ExpiresAt  *time.Time
}

type TaskUpdateStatusCmd struct {
	ID     uuid.UUID
	Status TaskStatus
}

type TaskUpdateForRetryCmd struct {
	ID           uuid.UUID
	Status       TaskStatus
	Retries      int
	NextRunAt    time.Time
	LastErrorMsg string
}
