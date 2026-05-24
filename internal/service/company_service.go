package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
)

type CompanyService struct {
	companyRepo repository.CompanyRepository
	userRepo    repository.UserRepository
	groupRepo   repository.GroupRepository
}

func NewCompanyService(companyRepo repository.CompanyRepository, userRepo repository.UserRepository, groupRepo repository.GroupRepository) *CompanyService {
	return &CompanyService{
		companyRepo: companyRepo,
		userRepo:    userRepo,
		groupRepo:   groupRepo,
	}
}

func (s *CompanyService) CreateCompany(ctx context.Context, name string, ownerID uuid.UUID) (*domain.Company, error) {
	user, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if user.CompanyID != nil {
		return nil, domain.ErrAlreadyInCompany
	}

	now := time.Now()
	company := &domain.Company{
		ID:        uuid.New(),
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.companyRepo.Create(ctx, company); err != nil {
		return nil, err
	}

	member := &domain.CompanyMember{
		ID:        uuid.New(),
		CompanyID: company.ID,
		UserID:    ownerID,
		Role:      domain.RoleOwner,
		JoinedAt:  now,
	}
	if err := s.companyRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	user.CompanyID = &company.ID
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Create default project group
	defaultGroup := &domain.ProjectGroup{
		ID:          uuid.New(),
		CompanyID:   company.ID,
		Name:        "General",
		Description: "Default project group",
		Color:       "#6366f1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.groupRepo.Create(ctx, defaultGroup); err != nil {
		return nil, err
	}

	return company, nil
}

func (s *CompanyService) GetCompany(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	return s.companyRepo.GetByID(ctx, id)
}

func (s *CompanyService) GetMembers(ctx context.Context, companyID uuid.UUID) ([]*domain.CompanyMember, error) {
	return s.companyRepo.GetMembers(ctx, companyID)
}

func (s *CompanyService) CreateInviteLink(ctx context.Context, companyID, invitedBy uuid.UUID) (*domain.Invite, error) {
	token, err := GenerateInviteToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	invite := &domain.Invite{
		ID:        uuid.New(),
		CompanyID: companyID,
		Token:     token,
		InvitedBy: invitedBy,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
	}

	if err := s.companyRepo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *CompanyService) CreateEmailInvite(ctx context.Context, companyID, invitedBy uuid.UUID, email string) (*domain.Invite, error) {
	token, err := GenerateInviteToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	invite := &domain.Invite{
		ID:        uuid.New(),
		CompanyID: companyID,
		Email:     email,
		Token:     token,
		InvitedBy: invitedBy,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
	}

	if err := s.companyRepo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *CompanyService) AcceptInvite(ctx context.Context, userID uuid.UUID, token string) (*domain.Company, error) {
	invite, err := s.companyRepo.GetInviteByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if invite.Used {
		return nil, domain.ErrInviteUsed
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, domain.ErrInviteExpired
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.CompanyID != nil {
		return nil, domain.ErrAlreadyInCompany
	}

	member := &domain.CompanyMember{
		ID:        uuid.New(),
		CompanyID: invite.CompanyID,
		UserID:    userID,
		Role:      domain.RoleMember,
		JoinedAt:  time.Now(),
	}

	if err := s.companyRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	user.CompanyID = &invite.CompanyID
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	if err := s.companyRepo.MarkInviteUsed(ctx, invite.ID); err != nil {
		return nil, err
	}

	return s.companyRepo.GetByID(ctx, invite.CompanyID)
}

func (s *CompanyService) GetInvites(ctx context.Context, companyID uuid.UUID) ([]*domain.Invite, error) {
	return s.companyRepo.GetInvitesByCompany(ctx, companyID)
}

func (s *CompanyService) RemoveMember(ctx context.Context, companyID, userID, requesterID uuid.UUID) error {
	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}
	if company.OwnerID != requesterID {
		return domain.ErrForbidden
	}
	if userID == requesterID {
		return domain.ErrValidation
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.CompanyID = nil
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return s.companyRepo.RemoveMember(ctx, companyID, userID)
}
