package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskRun struct {
	ID         uuid.UUID
	TaskID     uuid.UUID
	StartedAt  time.Time
	FinishedAt *time.Time
	Status     TaskStatus
	Error      string
	Output     string
	WorkerID   string
	DurationMs int64
}
