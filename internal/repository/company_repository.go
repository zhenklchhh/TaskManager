package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type CompanyRepository interface {
	Create(ctx context.Context, company *domain.Company) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Company, error)
	Update(ctx context.Context, company *domain.Company) error
	AddMember(ctx context.Context, member *domain.CompanyMember) error
	GetMembers(ctx context.Context, companyID uuid.UUID) ([]*domain.CompanyMember, error)
	RemoveMember(ctx context.Context, companyID, userID uuid.UUID) error
	CreateInvite(ctx context.Context, invite *domain.Invite) error
	GetInviteByToken(ctx context.Context, token string) (*domain.Invite, error)
	MarkInviteUsed(ctx context.Context, id uuid.UUID) error
	GetInvitesByCompany(ctx context.Context, companyID uuid.UUID) ([]*domain.Invite, error)
}
