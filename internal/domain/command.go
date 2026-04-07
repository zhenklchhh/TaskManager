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
	Priority   *int
	ExpiresAt  *time.Time
}

type TaskUpdateStatusCmd struct {
	ID           uuid.UUID
	Status       TaskStatus
	LastErrorMsg string
}

type TaskUpdateForRetryCmd struct {
	ID           uuid.UUID
	Status       TaskStatus
	Retries      int
	NextRunAt    time.Time
	LastErrorMsg string
}

type BatchCreateCmd struct {
	Tasks []TaskCreateCmd
}

type BatchCancelCmd struct {
	IDs []uuid.UUID
}

type BatchUpdatePriorityCmd struct {
	IDs      []uuid.UUID
	Priority int
}

type TaskFilter struct {
	Status      *TaskStatus
	Type        *string
	PriorityMin *int
	PriorityMax *int
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Limit       int
	Offset      int
}
