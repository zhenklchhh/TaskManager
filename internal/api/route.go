package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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

func Routes(h *Handler, healthChecker *HealthChecker, dashboardHandler *DashboardHandler,
	batchHandler *BatchHandler, depHandler *DependencyHandler, notifHandler *NotificationHandler) chi.Router {
	r := chi.NewRouter()

	// Add CORS middleware
	r.Use(corsMiddleware)
	
	// Add metrics middleware
	r.Use(MetricsMiddleware)
	
	// Health check endpoints
	r.Get("/health", healthChecker.Check)
	r.Get("/ready", healthChecker.Ready)
	
	// Metrics endpoint
	r.Handle("/metrics", MetricsHandler())
	
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router){
			r.Post("/tasks", h.CreateTask)
			r.Get("/tasks/{id}", h.GetTaskById)
			
			// Batch endpoints
			r.Post("/tasks/batch", batchHandler.BatchCreate)
			r.Post("/tasks/batch/cancel", batchHandler.BatchCancel)
			r.Put("/tasks/batch/priority", batchHandler.BatchUpdatePriority)
			
			// Dependency endpoints
			r.Post("/tasks/dependencies", depHandler.AddDependency)
			r.Get("/tasks/{id}/dependencies", depHandler.GetDependencies)
			r.Get("/tasks/{id}/dependents", depHandler.GetDependents)
			r.Get("/tasks/{id}/children", depHandler.GetChildTasks)
			r.Delete("/tasks/dependencies/{dep_id}", depHandler.RemoveDependency)
			
			// Notification endpoints
			r.Post("/notifications/config", notifHandler.CreateConfig)
			
			// Dashboard endpoints
			r.Get("/dashboard/stats", dashboardHandler.GetStats)
			r.Get("/dashboard/tasks", dashboardHandler.GetTasks)
			r.Get("/dashboard/tasks/filter", dashboardHandler.GetTasksFiltered)
		})
	})
	return r
}
