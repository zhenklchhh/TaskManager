package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type GroupService struct {
	groupRepo repository.GroupRepository
}

func NewGroupService(groupRepo repository.GroupRepository) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

func (s *GroupService) CreateGroup(ctx context.Context, companyID uuid.UUID, name, description, color string) (*domain.ProjectGroup, error) {
	if color == "" {
		color = "#6366f1"
	}
	now := time.Now()
	group := &domain.ProjectGroup{
		ID:          uuid.New(),
		CompanyID:   companyID,
		Name:        name,
		Description: description,
		Color:       color,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) GetGroupsByCompany(ctx context.Context, companyID uuid.UUID) ([]*domain.ProjectGroup, error) {
	return s.groupRepo.GetByCompany(ctx, companyID)
}

func (s *GroupService) GetGroup(ctx context.Context, id uuid.UUID) (*domain.ProjectGroup, error) {
	return s.groupRepo.GetByID(ctx, id)
}

func (s *GroupService) UpdateGroup(ctx context.Context, id uuid.UUID, name, description, color string) (*domain.ProjectGroup, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		group.Name = name
	}
	if description != "" {
		group.Description = description
	}
	if color != "" {
		group.Color = color
	}
	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return s.groupRepo.Delete(ctx, id)
}
