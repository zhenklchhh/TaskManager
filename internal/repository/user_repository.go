package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByOAuth(ctx context.Context, provider, oauthID string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}
