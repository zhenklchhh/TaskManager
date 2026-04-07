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
}
