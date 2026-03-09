package api

import (
	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler, healthChecker *HealthChecker) chi.Router {
	r := chi.NewRouter()
	
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
		})
	})
	return r
}
