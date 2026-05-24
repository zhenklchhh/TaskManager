package domain

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID        uuid.UUID
	Name      string
	OwnerID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleMember MemberRole = "member"
)

type CompanyMember struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	UserID    uuid.UUID
	Role      MemberRole
	JoinedAt  time.Time
	UserName  string
	UserEmail string
}
