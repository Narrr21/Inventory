package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"my-backend/internal/models"
	"my-backend/internal/repository"
	"my-backend/internal/response"

	"github.com/go-chi/chi/v5"
)

type ItemHandler struct {
	repo *repository.ItemRepository
}

func NewItemHandler(repo *repository.ItemRepository) *ItemHandler {
	return &ItemHandler{repo: repo}
}

func decodeBody(r *http.Request) map[string]interface{} {
	var body map[string]interface{}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return body
}

func decodeItemBody(r *http.Request) models.Item {
	var item models.Item
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&item)
	}
	return item
}

// CreateItem: POST /api/v1/items
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	item := decodeItemBody(r)
	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create item", nil)
		return
	}
	response.OK(w, http.StatusCreated, created, nil)
}

// ListItems: GET /api/v1/items
func (h *ItemHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filters := make(map[string]string)
	for _, field := range []string{
		"jenisProduct", "proyek", "status", "lokasi",
		"serialNumber", "name", "account", "ipAddress",
		"anydesk", "rustdesk", "licenseWindows", "licenseOffice", "q",
	} {
		if v := q.Get(field); v != "" {
			filters[field] = v
		}
	}

	result, err := h.repo.List(r.Context(), repository.ListParams{
		Page:      page,
		Limit:     limit,
		SortBy:    q.Get("sortBy"),
		SortOrder: q.Get("sortOrder"),
		Filters:   filters,
	})
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list items", nil)
		return
	}

	items := make([]models.ItemPublic, 0, len(result.Items))
	for _, it := range result.Items {
		items = append(items, it.Public())
	}
	meta := map[string]interface{}{
		"total":      result.Total,
		"page":       result.Page,
		"limit":      result.Limit,
		"totalPages": result.TotalPages,
	}
	response.OK(w, http.StatusOK, items, meta)
}

// GetItem: GET /api/v1/items/{id}
func (h *ItemHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.repo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Item not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get item", nil)
		return
	}
	response.OK(w, http.StatusOK, item, nil)
}

// UpdateItem: PATCH /api/v1/items/{id}
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	patch := decodeBody(r)
	item, err := h.repo.Update(r.Context(), id, patch)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Item not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update item", nil)
		return
	}
	response.OK(w, http.StatusOK, item, nil)
}

// DeleteItem: DELETE /api/v1/items/{id}
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.repo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Item not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete item", nil)
		return
	}
	response.OK(w, http.StatusOK, map[string]string{"_id": id}, nil)
}

// FilterOptions: GET /api/v1/items/filter-options
func (h *ItemHandler) FilterOptions(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.FilterOptions(r.Context(), "proyek", "jenisProduct", "lokasi", "status")
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load filter options", nil)
		return
	}
	response.OK(w, http.StatusOK, data, nil)
}

func (h *ItemHandler) distinctFieldValues(w http.ResponseWriter, r *http.Request, field string) {
	values, err := h.repo.Distinct(r.Context(), field)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load filter options", nil)
		return
	}
	response.OK(w, http.StatusOK, values, nil)
}

// FilterOptionsLokasi: GET /api/v1/items/filter-options/lokasi
func (h *ItemHandler) FilterOptionsLokasi(w http.ResponseWriter, r *http.Request) {
	h.distinctFieldValues(w, r, "lokasi")
}

// FilterOptionsProyek: GET /api/v1/items/filter-options/proyek
func (h *ItemHandler) FilterOptionsProyek(w http.ResponseWriter, r *http.Request) {
	h.distinctFieldValues(w, r, "proyek")
}

// FilterOptionsJenisProduct: GET /api/v1/items/filter-options/jenisProduct
func (h *ItemHandler) FilterOptionsJenisProduct(w http.ResponseWriter, r *http.Request) {
	h.distinctFieldValues(w, r, "jenisProduct")
}

// Stats: GET /api/v1/items/stats
func (h *ItemHandler) Stats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	total, err := h.repo.Count(ctx)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load stats", nil)
		return
	}
	byStatus, err := h.repo.CountBy(ctx, "status")
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load stats", nil)
		return
	}
	byJenisProduct, err := h.repo.CountBy(ctx, "jenisProduct")
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load stats", nil)
		return
	}
	byProyek, err := h.repo.CountBy(ctx, "proyek")
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load stats", nil)
		return
	}
	recent, err := h.repo.RecentlyAdded(ctx, 5)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load stats", nil)
		return
	}

	recentPublic := make([]models.ItemPublic, 0, len(recent))
	for _, it := range recent {
		recentPublic = append(recentPublic, it.Public())
	}

	data := map[string]interface{}{
		"totalItems":     total,
		"byStatus":       byStatus,
		"byJenisProduct": byJenisProduct,
		"byProyek":       byProyek,
		"recentlyAdded":  recentPublic,
	}
	response.OK(w, http.StatusOK, data, nil)
}

// ImportItems: POST /api/v1/items/import
func (h *ItemHandler) ImportItems(w http.ResponseWriter, r *http.Request) {
	data := map[string]int{"inserted": 0, "updated": 0, "failed": 0}
	response.OKWithErrors(w, http.StatusOK, data, []map[string]interface{}{})
}

// ExportItems: GET /api/v1/items/export
func (h *ItemHandler) ExportItems(w http.ResponseWriter, r *http.Request) {
	response.Err(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Export is not implemented in this phase", nil)
}
