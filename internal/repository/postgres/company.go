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

type PostgresCompanyRepository struct {
	pool *pgxpool.Pool
}

func NewCompanyRepository(pool *pgxpool.Pool) repository.CompanyRepository {
	return &PostgresCompanyRepository{pool: pool}
}

func (r *PostgresCompanyRepository) Create(ctx context.Context, company *domain.Company) error {
	const q = `
		INSERT INTO companies (id, name, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, q,
		company.ID, company.Name, company.OwnerID, company.CreatedAt, company.UpdatedAt,
	)
	return err
}

func (r *PostgresCompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	const q = `
		SELECT id, name, owner_id, created_at, updated_at
		FROM companies WHERE id = $1
	`
	var c domain.Company
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&c.ID, &c.Name, &c.OwnerID, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *PostgresCompanyRepository) Update(ctx context.Context, company *domain.Company) error {
	const q = `UPDATE companies SET name = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, company.ID, company.Name)
	return err
}

func (r *PostgresCompanyRepository) AddMember(ctx context.Context, member *domain.CompanyMember) error {
	const q = `
		INSERT INTO company_members (id, company_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, q,
		member.ID, member.CompanyID, member.UserID, member.Role, member.JoinedAt,
	)
	return err
}

func (r *PostgresCompanyRepository) GetMembers(ctx context.Context, companyID uuid.UUID) ([]*domain.CompanyMember, error) {
	const q = `
		SELECT cm.id, cm.company_id, cm.user_id, cm.role, cm.joined_at, u.name, u.email
		FROM company_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.company_id = $1
		ORDER BY cm.joined_at
	`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.CompanyMember
	for rows.Next() {
		var m domain.CompanyMember
		if err := rows.Scan(&m.ID, &m.CompanyID, &m.UserID, &m.Role, &m.JoinedAt, &m.UserName, &m.UserEmail); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, nil
}

func (r *PostgresCompanyRepository) RemoveMember(ctx context.Context, companyID, userID uuid.UUID) error {
	const q = `DELETE FROM company_members WHERE company_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, q, companyID, userID)
	return err
}

func (r *PostgresCompanyRepository) CreateInvite(ctx context.Context, invite *domain.Invite) error {
	const q = `
		INSERT INTO invites (id, company_id, email, token, invited_by, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, q,
		invite.ID, invite.CompanyID, nullString(invite.Email), invite.Token,
		invite.InvitedBy, invite.ExpiresAt, invite.CreatedAt,
	)
	return err
}

func (r *PostgresCompanyRepository) GetInviteByToken(ctx context.Context, token string) (*domain.Invite, error) {
	const q = `
		SELECT id, company_id, COALESCE(email, '') AS email, token, invited_by, expires_at, used, created_at
		FROM invites WHERE token = $1
	`
	var i domain.Invite
	err := r.pool.QueryRow(ctx, q, token).Scan(
		&i.ID, &i.CompanyID, &i.Email, &i.Token, &i.InvitedBy, &i.ExpiresAt, &i.Used, &i.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInviteNotFound
		}
		return nil, err
	}
	return &i, nil
}

func (r *PostgresCompanyRepository) MarkInviteUsed(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE invites SET used = TRUE WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}

func (r *PostgresCompanyRepository) GetInvitesByCompany(ctx context.Context, companyID uuid.UUID) ([]*domain.Invite, error) {
	const q = `
		SELECT id, company_id, COALESCE(email, '') AS email, token, invited_by, expires_at, used, created_at
		FROM invites WHERE company_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*domain.Invite
	for rows.Next() {
		var i domain.Invite
		if err := rows.Scan(&i.ID, &i.CompanyID, &i.Email, &i.Token, &i.InvitedBy, &i.ExpiresAt, &i.Used, &i.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, &i)
	}
	return invites, nil
}
