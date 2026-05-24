package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type RouteParams struct {
	Handler            *Handler
	HealthChecker      *HealthChecker
	DashboardHandler   *DashboardHandler
	BatchHandler       *BatchHandler
	DepHandler         *DependencyHandler
	NotifHandler       *NotificationHandler
	AuthHandler        *AuthHandler
	CompanyHandler     *CompanyHandler
	GroupHandler       *GroupHandler
	AuthService        *service.AuthService
}

func Routes(p RouteParams) chi.Router {
	r := chi.NewRouter()

	// Add CORS middleware
	r.Use(corsMiddleware)
	
	// Add metrics middleware
	r.Use(MetricsMiddleware)
	
	// Health check endpoints
	r.Get("/health", p.HealthChecker.Check)
	r.Get("/ready", p.HealthChecker.Ready)
	
	// Metrics endpoint
	r.Handle("/metrics", MetricsHandler())
	
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Public auth endpoints
			r.Post("/auth/register", p.AuthHandler.Register)
			r.Post("/auth/login", p.AuthHandler.Login)
			r.Post("/auth/google/callback", p.AuthHandler.GoogleCallback)
			r.Post("/auth/github/callback", p.AuthHandler.GitHubCallback)
			r.Get("/auth/google/url", p.AuthHandler.GetGoogleAuthURL)
			r.Get("/auth/github/url", p.AuthHandler.GetGitHubAuthURL)

			// Protected routes
			r.Group(func(r chi.Router) {
				r.Use(AuthMiddleware(p.AuthService))

				// User profile
				r.Get("/auth/me", p.AuthHandler.Me)
				r.Post("/auth/refresh", p.AuthHandler.RefreshToken)

				// Company management
				r.Post("/companies", p.CompanyHandler.Create)
				r.Post("/companies/join", p.CompanyHandler.AcceptInvite)

				// Routes requiring company membership
				r.Group(func(r chi.Router) {
					r.Use(RequireCompany)

					r.Get("/companies/current", p.CompanyHandler.Get)
					r.Get("/companies/members", p.CompanyHandler.GetMembers)
					r.Post("/companies/members/remove", p.CompanyHandler.RemoveMember)
					r.Post("/companies/invites/link", p.CompanyHandler.CreateInviteLink)
					r.Post("/companies/invites/email", p.CompanyHandler.CreateEmailInvite)
					r.Get("/companies/invites", p.CompanyHandler.GetInvites)

					// Project groups
					r.Post("/groups", p.GroupHandler.Create)
					r.Get("/groups", p.GroupHandler.List)
					r.Get("/groups/{id}", p.GroupHandler.Get)
					r.Put("/groups/{id}", p.GroupHandler.Update)
					r.Delete("/groups/{id}", p.GroupHandler.Delete)

					// Tasks (now company-scoped)
					r.Post("/tasks", p.Handler.CreateTask)
					r.Get("/tasks/{id}", p.Handler.GetTaskById)
					r.Patch("/tasks/{id}", p.Handler.UpdateTask)
					r.Delete("/tasks/{id}", p.Handler.DeleteTask)
					r.Post("/tasks/{id}/cancel", p.Handler.CancelTask)
					r.Post("/tasks/{id}/retry", p.Handler.RetryTask)
					r.Get("/tasks/{id}/logs", p.Handler.GetTaskLogs)

					// Batch endpoints
					r.Post("/tasks/batch", p.BatchHandler.BatchCreate)
					r.Post("/tasks/batch/cancel", p.BatchHandler.BatchCancel)
					r.Put("/tasks/batch/priority", p.BatchHandler.BatchUpdatePriority)

					// Dependency endpoints
					r.Post("/tasks/dependencies", p.DepHandler.AddDependency)
					r.Get("/tasks/{id}/dependencies", p.DepHandler.GetDependencies)
					r.Get("/tasks/{id}/dependents", p.DepHandler.GetDependents)
					r.Get("/tasks/{id}/children", p.DepHandler.GetChildTasks)
					r.Delete("/tasks/dependencies/{dep_id}", p.DepHandler.RemoveDependency)

					// Notification endpoints
					r.Post("/notifications/config", p.NotifHandler.CreateConfig)

					// Dashboard endpoints
					r.Get("/dashboard/stats", p.DashboardHandler.GetStats)
					r.Get("/dashboard/tasks", p.DashboardHandler.GetTasks)
					r.Get("/dashboard/tasks/filter", p.DashboardHandler.GetTasksFiltered)
				})
			})
		})
	})
	return r
}
