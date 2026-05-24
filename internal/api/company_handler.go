package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type CompanyHandler struct {
	companyService *service.CompanyService
}

func NewCompanyHandler(companyService *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{companyService: companyService}
}

type CreateCompanyRequest struct {
	Name string `json:"name" validate:"required"`
}

type CompanyResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerID   string `json:"owner_id"`
	CreatedAt string `json:"created_at"`
}

type MemberResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

type InviteResponse struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	Email     string `json:"email,omitempty"`
	ExpiresAt string `json:"expires_at"`
	Used      bool   `json:"used"`
	Link      string `json:"link"`
}

type CreateEmailInviteRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type AcceptInviteRequest struct {
	Token string `json:"token" validate:"required"`
}

type RemoveMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

func (h *CompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := requestValidator.Struct(req); err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r.Context())
	company, err := h.companyService.CreateCompany(r.Context(), req.Name, userID)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapCompanyResp(company))
}

func (h *CompanyHandler) Get(w http.ResponseWriter, r *http.Request) {
	companyID := GetCompanyID(r.Context())
	if companyID == nil {
		http.Error(w, "no company", http.StatusNotFound)
		return
	}
	company, err := h.companyService.GetCompany(r.Context(), *companyID)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapCompanyResp(company))
}

func (h *CompanyHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	companyID := GetCompanyID(r.Context())
	if companyID == nil {
		http.Error(w, "no company", http.StatusNotFound)
		return
	}
	members, err := h.companyService.GetMembers(r.Context(), *companyID)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	resp := make([]MemberResponse, len(members))
	for i, m := range members {
		resp[i] = MemberResponse{
			ID:       m.ID.String(),
			UserID:   m.UserID.String(),
			Name:     m.UserName,
			Email:    m.UserEmail,
			Role:     string(m.Role),
			JoinedAt: m.JoinedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CompanyHandler) CreateInviteLink(w http.ResponseWriter, r *http.Request) {
	companyID := GetCompanyID(r.Context())
	if companyID == nil {
		http.Error(w, "no company", http.StatusNotFound)
		return
	}

	userID := GetUserID(r.Context())
	invite, err := h.companyService.CreateInviteLink(r.Context(), *companyID, userID)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapInviteResp(invite))
}

func (h *CompanyHandler) CreateEmailInvite(w http.ResponseWriter, r *http.Request) {
	var req CreateEmailInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := requestValidator.Struct(req); err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	companyID := GetCompanyID(r.Context())
	if companyID == nil {
		http.Error(w, "no company", http.StatusNotFound)
		return
	}

	userID := GetUserID(r.Context())
	invite, err := h.companyService.CreateEmailInvite(r.Context(), *companyID, userID, req.Email)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapInviteResp(invite))
}

func (h *CompanyHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	userID := GetUserID(r.Context())
	company, err := h.companyService.AcceptInvite(r.Context(), userID, req.Token)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapCompanyResp(company))
}

func (h *CompanyHandler) GetInvites(w http.ResponseWriter, r *http.Request) {
	companyID := GetCompanyID(r.Context())
	if companyID == nil {
		http.Error(w, "no company", http.StatusNotFound)
		return
	}

	invites, err := h.companyService.GetInvites(r.Context(), *companyID)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	resp := make([]InviteResponse, len(invites))
	for i, inv := range invites {
		resp[i] = mapInviteResp(inv)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CompanyHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	var req RemoveMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	companyID := GetCompanyID(r.Context())
	if companyID == nil {
		http.Error(w, "no company", http.StatusNotFound)
		return
	}

	targetID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	requesterID := GetUserID(r.Context())
	if err := h.companyService.RemoveMember(r.Context(), *companyID, targetID, requesterID); err != nil {
		handleAuthError(err, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapCompanyResp(c *domain.Company) CompanyResponse {
	return CompanyResponse{
		ID:        c.ID.String(),
		Name:      c.Name,
		OwnerID:   c.OwnerID.String(),
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

func mapInviteResp(i *domain.Invite) InviteResponse {
	return InviteResponse{
		ID:        i.ID.String(),
		Token:     i.Token,
		Email:     i.Email,
		ExpiresAt: i.ExpiresAt.Format(time.RFC3339),
		Used:      i.Used,
		Link:      "/invite/" + i.Token,
	}
}
