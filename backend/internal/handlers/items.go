package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"my-backend/internal/mocks"
	"my-backend/internal/response"
)

func decodeBody(r *http.Request) map[string]interface{} {
	var body map[string]interface{}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return body
}

// CreateItem: POST /api/v1/items
func CreateItem(w http.ResponseWriter, r *http.Request) {
	body := decodeBody(r)
	item := mocks.MergeItem(mocks.Items[0], body)
	item.ID = "665f1a1a1a1a1a1a1a1a1a99"
	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	item.CreatedAt = now
	item.UpdatedAt = now
	response.OK(w, http.StatusCreated, item, nil)
}

// ListItems: GET /api/v1/items
func ListItems(w http.ResponseWriter, r *http.Request) {
	items := make([]mocks.ItemPublic, 0, len(mocks.Items))
	for _, it := range mocks.Items {
		items = append(items, it.Public())
	}
	meta := map[string]interface{}{
		"total":      len(mocks.Items),
		"page":       1,
		"limit":      20,
		"totalPages": 1,
	}
	response.OK(w, http.StatusOK, items, meta)
}

// GetItem: GET /api/v1/items/{id}
func GetItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "notfound" {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Item not found", nil)
		return
	}
	response.OK(w, http.StatusOK, mocks.Items[0], nil)
}

// UpdateItem: PATCH /api/v1/items/{id}
func UpdateItem(w http.ResponseWriter, r *http.Request) {
	body := decodeBody(r)
	item := mocks.MergeItem(mocks.Items[0], body)
	response.OK(w, http.StatusOK, item, nil)
}

// DeleteItem: DELETE /api/v1/items/{id}
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	response.OK(w, http.StatusOK, map[string]string{"_id": id}, nil)
}

// FilterOptions: GET /api/v1/items/filter-options
func FilterOptions(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"proyek":       []string{"ALPHA", "BETA", "GAMMA"},
		"jenisProduct": []string{"Laptop", "PC", "Monitor"},
		"lokasi":       []string{"Jakarta HQ", "Surabaya Branch"},
		"status":       []string{"Active", "Idle", "Maintenance"},
	}
	response.OK(w, http.StatusOK, data, nil)
}

// Stats: GET /api/v1/items/stats
func Stats(w http.ResponseWriter, r *http.Request) {
	recent := make([]mocks.ItemPublic, 0, len(mocks.Items))
	for _, it := range mocks.Items {
		recent = append(recent, it.Public())
	}
	data := map[string]interface{}{
		"totalItems": 143,
		"byStatus": []map[string]interface{}{
			{"_id": "Active", "count": 120},
			{"_id": "Maintenance", "count": 15},
		},
		"byJenisProduct": []map[string]interface{}{
			{"_id": "Laptop", "count": 80},
			{"_id": "PC", "count": 40},
		},
		"byProyek": []map[string]interface{}{
			{"_id": "ALPHA", "count": 30},
		},
		"recentlyAdded": recent,
	}
	response.OK(w, http.StatusOK, data, nil)
}

// ImportItems: POST /api/v1/items/import
func ImportItems(w http.ResponseWriter, r *http.Request) {
	data := map[string]int{"inserted": 0, "updated": 0, "failed": 0}
	response.OKWithErrors(w, http.StatusOK, data, []map[string]interface{}{})
}

// ExportItems: GET /api/v1/items/export
func ExportItems(w http.ResponseWriter, r *http.Request) {
	response.Err(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Export is not implemented in this phase", nil)
}
