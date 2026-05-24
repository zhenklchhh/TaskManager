package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/config"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    repository.UserRepository
	companyRepo repository.CompanyRepository
	authCfg     config.AuthConfig
}

func NewAuthService(userRepo repository.UserRepository, companyRepo repository.CompanyRepository, authCfg config.AuthConfig) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		companyRepo: companyRepo,
		authCfg:     authCfg,
	}
}

type JWTClaims struct {
	UserID    uuid.UUID  `json:"user_id"`
	Email     string     `json:"email"`
	CompanyID *uuid.UUID `json:"company_id,omitempty"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, string, error) {
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, "", domain.ErrEmailExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", domain.ErrInvalidPassword
		}
		return nil, "", err
	}

	if user.PasswordHash == "" {
		return nil, "", domain.ErrInvalidPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", domain.ErrInvalidPassword
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*domain.User, string, error) {
	tokenResp, err := exchangeGoogleCode(code, s.authCfg.GoogleClientID, s.authCfg.GoogleClientSecret, s.authCfg.OAuthCallbackURL+"/auth/google/callback")
	if err != nil {
		return nil, "", fmt.Errorf("google token exchange failed: %w", err)
	}

	userInfo, err := getGoogleUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, "", fmt.Errorf("google user info failed: %w", err)
	}

	return s.findOrCreateOAuthUser(ctx, "google", userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture)
}

func (s *AuthService) HandleGitHubCallback(ctx context.Context, code string) (*domain.User, string, error) {
	tokenResp, err := exchangeGitHubCode(code, s.authCfg.GitHubClientID, s.authCfg.GitHubClientSecret)
	if err != nil {
		return nil, "", fmt.Errorf("github token exchange failed: %w", err)
	}

	userInfo, err := getGitHubUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, "", fmt.Errorf("github user info failed: %w", err)
	}

	email := userInfo.Email
	if email == "" {
		email = fmt.Sprintf("%s@github.local", userInfo.Login)
	}

	return s.findOrCreateOAuthUser(ctx, "github", fmt.Sprintf("%d", userInfo.ID), email, userInfo.Name, userInfo.AvatarURL)
}

func (s *AuthService) findOrCreateOAuthUser(ctx context.Context, provider, oauthID, email, name, avatar string) (*domain.User, string, error) {
	user, err := s.userRepo.GetByOAuth(ctx, provider, oauthID)
	if err == nil {
		token, err := s.generateJWT(user)
		if err != nil {
			return nil, "", err
		}
		return user, token, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, "", err
	}

	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		existing.OAuthProvider = provider
		existing.OAuthID = oauthID
		if avatar != "" {
			existing.AvatarURL = avatar
		}
		if err := s.userRepo.Update(ctx, existing); err != nil {
			return nil, "", err
		}
		token, err := s.generateJWT(existing)
		if err != nil {
			return nil, "", err
		}
		return existing, token, nil
	}

	if name == "" {
		name = email
	}

	now := time.Now()
	user = &domain.User{
		ID:            uuid.New(),
		Email:         email,
		Name:          name,
		AvatarURL:     avatar,
		OAuthProvider: provider,
		OAuthID:       oauthID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrUnauthorized
		}
		return []byte(s.authCfg.JWTSecret), nil
	})
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	return claims, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return s.generateJWT(user)
}

func (s *AuthService) generateJWT(user *domain.User) (string, error) {
	claims := &JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		CompanyID: user.CompanyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.authCfg.JWTExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.authCfg.JWTSecret))
}

func (s *AuthService) GetGoogleAuthURL() string {
	return fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20email%%20profile",
		s.authCfg.GoogleClientID,
		s.authCfg.OAuthCallbackURL+"/auth/google/callback",
	)
}

func (s *AuthService) GetGitHubAuthURL() string {
	return fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email",
		s.authCfg.GitHubClientID,
		s.authCfg.OAuthCallbackURL+"/auth/github/callback",
	)
}

func GenerateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// OAuth helper types and functions

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type githubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func exchangeGoogleCode(code, clientID, clientSecret, redirectURI string) (*oauthTokenResponse, error) {
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", map[string][]string{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenResp oauthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access token from google")
	}
	return &tokenResp, nil
}

func getGoogleUserInfo(accessToken string) (*googleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func exchangeGitHubCode(code, clientID, clientSecret string) (*oauthTokenResponse, error) {
	req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	q := req.URL.Query()
	q.Add("client_id", clientID)
	q.Add("client_secret", clientSecret)
	q.Add("code", code)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenResp oauthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access token from github")
	}
	return &tokenResp, nil
}

func getGitHubUserInfo(accessToken string) (*githubUserInfo, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info githubUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
