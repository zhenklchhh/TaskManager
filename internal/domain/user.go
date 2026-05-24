package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Email         string
	PasswordHash  string
	Name          string
	AvatarURL     string
	OAuthProvider string
	OAuthID       string
	CompanyID     *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
