package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users (id, email, password_hash, name, avatar_url, oauth_provider, oauth_id, company_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, q,
		user.ID, user.Email, user.PasswordHash, user.Name, user.AvatarURL,
		nullString(user.OAuthProvider), nullString(user.OAuthID),
		user.CompanyID, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, email, COALESCE(password_hash, '') AS password_hash, name,
			COALESCE(avatar_url, '') AS avatar_url,
			COALESCE(oauth_provider, '') AS oauth_provider,
			COALESCE(oauth_id, '') AS oauth_id,
			company_id, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name,
		&u.AvatarURL, &u.OAuthProvider, &u.OAuthID,
		&u.CompanyID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, COALESCE(password_hash, '') AS password_hash, name,
			COALESCE(avatar_url, '') AS avatar_url,
			COALESCE(oauth_provider, '') AS oauth_provider,
			COALESCE(oauth_id, '') AS oauth_id,
			company_id, created_at, updated_at
		FROM users WHERE email = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name,
		&u.AvatarURL, &u.OAuthProvider, &u.OAuthID,
		&u.CompanyID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByOAuth(ctx context.Context, provider, oauthID string) (*domain.User, error) {
	const q = `
		SELECT id, email, COALESCE(password_hash, '') AS password_hash, name,
			COALESCE(avatar_url, '') AS avatar_url,
			COALESCE(oauth_provider, '') AS oauth_provider,
			COALESCE(oauth_id, '') AS oauth_id,
			company_id, created_at, updated_at
		FROM users WHERE oauth_provider = $1 AND oauth_id = $2
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, provider, oauthID).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name,
		&u.AvatarURL, &u.OAuthProvider, &u.OAuthID,
		&u.CompanyID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	const q = `
		UPDATE users SET email = $2, name = $3, avatar_url = $4, company_id = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, q, user.ID, user.Email, user.Name, user.AvatarURL, user.CompanyID)
	return err
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
