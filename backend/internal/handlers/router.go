package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(items *ItemHandler, projects *ProjectHandler, itemTypes *ItemTypeHandler, analytics *AnalyticsHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders: []string{"Accept", "Content-Type"},
		MaxAge:         300,
	}))

	r.Route("/api/v1/items", func(r chi.Router) {
		r.Post("/", items.CreateItem)
		r.Get("/", items.ListItems)
		r.Get("/filter-options", items.FilterOptions)
		r.Get("/filter-options/jenis", items.FilterOptionsJenis)
		r.Post("/import", items.ImportItems)
		r.Get("/export", items.ExportItems)
		r.Get("/{id}", items.GetItem)
		r.Patch("/{id}", items.UpdateItem)
		r.Delete("/{id}", items.DeleteItem)
	})

	r.Route("/api/v1/projects", func(r chi.Router) {
		r.Post("/", projects.CreateProject)
		r.Get("/", projects.ListProjects)
		r.Get("/{id}", projects.GetProject)
		r.Patch("/{id}", projects.UpdateProject)
		r.Delete("/{id}", projects.DeleteProject)
	})

	r.Route("/api/v1/item-types", func(r chi.Router) {
		r.Post("/", itemTypes.CreateItemType)
		r.Get("/", itemTypes.ListItemTypes)
		r.Delete("/{id}", itemTypes.DeleteItemType)
	})

	r.Route("/api/v1/analytics", func(r chi.Router) {
		r.Get("/summary", analytics.Summary)
		r.Get("/map", analytics.Map)
		r.Get("/timeline", analytics.Timeline)
	})

	return r
}
