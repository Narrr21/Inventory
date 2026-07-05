package handlers

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(items *ItemHandler) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/v1/items", func(r chi.Router) {
		r.Post("/", items.CreateItem)
		r.Get("/", items.ListItems)
		r.Get("/filter-options", items.FilterOptions)
		r.Get("/stats", items.Stats)
		r.Post("/import", items.ImportItems)
		r.Get("/export", items.ExportItems)
		r.Get("/{id}", items.GetItem)
		r.Patch("/{id}", items.UpdateItem)
		r.Delete("/{id}", items.DeleteItem)
	})

	return r
}
