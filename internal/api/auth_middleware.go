package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type contextKey string

const (
	ContextUserID    contextKey = "user_id"
	ContextCompanyID contextKey = "company_id"
)

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := authService.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			if claims.CompanyID != nil {
				ctx = context.WithValue(ctx, ContextCompanyID, *claims.CompanyID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ContextUserID).(uuid.UUID)
	return id
}

func GetCompanyID(ctx context.Context) *uuid.UUID {
	id, ok := ctx.Value(ContextCompanyID).(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}

func RequireCompany(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		companyID := GetCompanyID(r.Context())
		if companyID == nil {
			http.Error(w, "company membership required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
