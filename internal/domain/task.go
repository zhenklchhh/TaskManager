package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusScheduled TaskStatus = "scheduled"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID           uuid.UUID
	ParentID     *uuid.UUID
	Title        string
	Type         string
	LastErrorMsg string
	Payload      []byte
	CronExpr     string
	NextRunAt    time.Time
	Status       TaskStatus
	RetryCount   int
	MaxRetries   int
	Priority     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ExpiresAt    *time.Time
	CompanyID    *uuid.UUID
	GroupID      *uuid.UUID
	CreatedBy    *uuid.UUID
	AssignedTo   *uuid.UUID
	DeletedAt    *time.Time
}
