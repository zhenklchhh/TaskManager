package http

import "github.com/go-chi/chi"

func Routes(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router){
			r.Post("/tasks", h.CreateTask)
			r.Get("/tasks/{id}", h.GetTaskById)
		})
	})
	return r
}