package handlers

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(items *ItemHandler, projects *ProjectHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Timeout(10 * time.Second))

	r.Route("/api/v1/items", func(r chi.Router) {
		r.Post("/", items.CreateItem)
		r.Get("/", items.ListItems)
		r.Get("/filter-options", items.FilterOptions)
		r.Get("/filter-options/jenis", items.FilterOptionsJenis)
		r.Get("/stats", items.Stats)
		r.Post("/import", items.ImportItems)
		r.Get("/export", items.ExportItems)
		r.Get("/{id}", items.GetItem)
		r.Patch("/{id}", items.UpdateItem)
		r.Delete("/{id}", items.DeleteItem)
	})

	r.Route("/api/v1/projects", func(r chi.Router) {
		r.Post("/", projects.CreateProject)
		r.Get("/", projects.ListProjects)
	})

	return r
}
