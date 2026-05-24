package domain

import (
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Email     string
	Token     string
	InvitedBy uuid.UUID
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}
