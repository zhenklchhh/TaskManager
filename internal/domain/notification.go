package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationEmail   NotificationType = "email"
	NotificationWebhook NotificationType = "webhook"
)

type NotificationEvent string

const (
	EventTaskCompleted NotificationEvent = "task_completed"
	EventTaskFailed    NotificationEvent = "task_failed"
	EventStatusChanged NotificationEvent = "status_changed"
)

type NotificationConfig struct {
	ID        uuid.UUID
	TaskID    *uuid.UUID
	Type      NotificationType
	Event     NotificationEvent
	Target    string
	CreatedAt time.Time
}

type NotificationLog struct {
	ID               uuid.UUID
	ConfigID         uuid.UUID
	TaskID           uuid.UUID
	Event            NotificationEvent
	Status           string
	Attempts         int
	MaxAttempts      int
	LastError        string
	NextRetryAt      *time.Time
	CreatedAt        time.Time
	LastAttemptAt    *time.Time
}
