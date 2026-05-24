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
	CompanyID  *uuid.UUID
	GroupID    *uuid.UUID
	CreatedBy  *uuid.UUID
	AssignedTo *uuid.UUID
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

type TaskUpdateCmd struct {
	ID         uuid.UUID
	Title      *string
	Payload    *string
	Priority   *int
	CronExpr   *string
	MaxRetries *int
	NextRunAt  *time.Time
	GroupID    *uuid.UUID
	AssignedTo *uuid.UUID
	ClearGroup bool
}

type TaskFilter struct {
	Status      *TaskStatus
	Type        *string
	PriorityMin *int
	PriorityMax *int
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	CompanyID   *uuid.UUID
	GroupID     *uuid.UUID
	AssignedTo  *uuid.UUID
	Limit       int
	Offset      int
}
