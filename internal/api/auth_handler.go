package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type OAuthCallbackRequest struct {
	Code string `json:"code" validate:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	CompanyID *string `json:"company_id,omitempty"`
}

func mapUserResp(u *domain.User) UserResponse {
	resp := UserResponse{
		ID:    u.ID.String(),
		Email: u.Email,
		Name:  u.Name,
	}
	if u.AvatarURL != "" {
		resp.AvatarURL = u.AvatarURL
	}
	if u.CompanyID != nil {
		s := u.CompanyID.String()
		resp.CompanyID = &s
	}
	return resp
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := requestValidator.Struct(req); err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	user, token, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: mapUserResp(user)})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := requestValidator.Struct(req); err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	user, token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: mapUserResp(user)})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		handleAuthError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapUserResp(user))
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	var req OAuthCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, token, err := h.authService.HandleGoogleCallback(r.Context(), req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: mapUserResp(user)})
}

func (h *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	var req OAuthCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, token, err := h.authService.HandleGitHubCallback(r.Context(), req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: mapUserResp(user)})
}

func (h *AuthHandler) GetGoogleAuthURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": h.authService.GetGoogleAuthURL()})
}

func (h *AuthHandler) GetGitHubAuthURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": h.authService.GetGitHubAuthURL()})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		handleAuthError(err, w)
		return
	}
	token, err := h.authService.RefreshToken(r.Context(), userID)
	if err != nil {
		handleAuthError(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: mapUserResp(user)})
}

func handleAuthError(err error, w http.ResponseWriter) {
	switch {
	case errors.Is(err, domain.ErrEmailExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrInvalidPassword):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, domain.ErrUserNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrUnauthorized):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, domain.ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrAlreadyInCompany):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrInviteNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrInviteExpired):
		http.Error(w, err.Error(), http.StatusGone)
	case errors.Is(err, domain.ErrInviteUsed):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrCompanyNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
