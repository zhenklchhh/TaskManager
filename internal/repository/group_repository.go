package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type GroupRepository interface {
	Create(ctx context.Context, group *domain.ProjectGroup) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ProjectGroup, error)
	GetByCompany(ctx context.Context, companyID uuid.UUID) ([]*domain.ProjectGroup, error)
	Update(ctx context.Context, group *domain.ProjectGroup) error
	Delete(ctx context.Context, id uuid.UUID) error
}
