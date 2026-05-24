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

type PostgresGroupRepository struct {
	pool *pgxpool.Pool
}

func NewGroupRepository(pool *pgxpool.Pool) repository.GroupRepository {
	return &PostgresGroupRepository{pool: pool}
}

func (r *PostgresGroupRepository) Create(ctx context.Context, group *domain.ProjectGroup) error {
	const q = `
		INSERT INTO project_groups (id, company_id, name, description, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, q,
		group.ID, group.CompanyID, group.Name, group.Description, group.Color,
		group.CreatedAt, group.UpdatedAt,
	)
	return err
}

func (r *PostgresGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProjectGroup, error) {
	const q = `
		SELECT g.id, g.company_id, g.name, COALESCE(g.description, '') AS description, 
			g.color, g.created_at, g.updated_at,
			COUNT(t.id) AS task_count
		FROM project_groups g
		LEFT JOIN tasks t ON t.group_id = g.id AND t.deleted_at IS NULL
		WHERE g.id = $1
		GROUP BY g.id
	`
	var g domain.ProjectGroup
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&g.ID, &g.CompanyID, &g.Name, &g.Description,
		&g.Color, &g.CreatedAt, &g.UpdatedAt, &g.TaskCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *PostgresGroupRepository) GetByCompany(ctx context.Context, companyID uuid.UUID) ([]*domain.ProjectGroup, error) {
	const q = `
		SELECT g.id, g.company_id, g.name, COALESCE(g.description, '') AS description, 
			g.color, g.created_at, g.updated_at,
			COUNT(t.id) AS task_count
		FROM project_groups g
		LEFT JOIN tasks t ON t.group_id = g.id AND t.deleted_at IS NULL
		WHERE g.company_id = $1
		GROUP BY g.id
		ORDER BY g.name
	`
	rows, err := r.pool.Query(ctx, q, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.ProjectGroup
	for rows.Next() {
		var g domain.ProjectGroup
		if err := rows.Scan(&g.ID, &g.CompanyID, &g.Name, &g.Description,
			&g.Color, &g.CreatedAt, &g.UpdatedAt, &g.TaskCount); err != nil {
			return nil, err
		}
		groups = append(groups, &g)
	}
	return groups, nil
}

func (r *PostgresGroupRepository) Update(ctx context.Context, group *domain.ProjectGroup) error {
	const q = `
		UPDATE project_groups SET name = $2, description = $3, color = $4, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, q, group.ID, group.Name, group.Description, group.Color)
	return err
}

func (r *PostgresGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM project_groups WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}
