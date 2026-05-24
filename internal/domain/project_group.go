package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectGroup struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Name        string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TaskCount   int
}
