package domain

import(
	"errors"
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

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidCron = errors.New("invalid cron expression")
	ErrValidation = errors.New("invalid parameters")
)

type Task struct {
	ID uuid.UUID
	Title string
	Type string
	Payload []byte
	CronExpr string
	NextRunAt time.Time
	Status TaskStatus
	RetryCount int
	MaxRetries int	
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TaskRun struct {
	ID uuid.UUID
	TaskID uuid.UUID
	StartedAt time.Time
	FinishedAt time.Time
	Status TaskStatus
	Error error
}