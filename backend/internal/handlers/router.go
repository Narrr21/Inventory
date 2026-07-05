package handlers

import "net/http"

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/items", CreateItem)
	mux.HandleFunc("GET /api/v1/items", ListItems)
	mux.HandleFunc("GET /api/v1/items/filter-options", FilterOptions)
	mux.HandleFunc("GET /api/v1/items/stats", Stats)
	mux.HandleFunc("POST /api/v1/items/import", ImportItems)
	mux.HandleFunc("GET /api/v1/items/export", ExportItems)
	mux.HandleFunc("GET /api/v1/items/{id}", GetItem)
	mux.HandleFunc("PATCH /api/v1/items/{id}", UpdateItem)
	mux.HandleFunc("DELETE /api/v1/items/{id}", DeleteItem)

	return mux
}
