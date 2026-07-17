package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"my-backend/internal/models"
	"my-backend/internal/repository"
	"my-backend/internal/response"

	"github.com/go-chi/chi/v5"
)

type ItemHandler struct {
	repo        *repository.ItemRepository
	projectRepo *repository.ProjectRepository
}

func NewItemHandler(repo *repository.ItemRepository, projectRepo *repository.ProjectRepository) *ItemHandler {
	return &ItemHandler{repo: repo, projectRepo: projectRepo}
}

func decodeBody(r *http.Request) map[string]interface{} {
	var body map[string]interface{}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return body
}

// knownItemFields are the top-level JSON keys the Item model declares.
var knownItemFields = map[string]bool{
	"jenis": true, "serialNumber": true, "nama": true, "idProyek": true,
	"credentials": true, "remoteInfo": true, "licenseWindows": true, "licenseOffice": true,
	"status": true, "deskripsi": true, "createdAt": true, "updatedAt": true,
}

// splitCustomAttributes separates raw's top-level keys into recognized Item
// fields (returned as-is in known) and everything else (returned in custom).
// Any customAttributes object the client sent explicitly is folded into
// custom too. This is what makes "add a new column out of nowhere" actually
// work: an unrecognized field is never silently dropped (as Create used to)
// or written straight into the document root ungoverned (as Update used to)
// — it always lands in the one place the schema reserves for ad-hoc data.
func splitCustomAttributes(raw map[string]interface{}) (known map[string]interface{}, custom map[string]interface{}) {
	known = make(map[string]interface{}, len(raw))
	custom = make(map[string]interface{})

	if explicit, ok := raw["customAttributes"].(map[string]interface{}); ok {
		for k, v := range explicit {
			custom[k] = v
		}
	}

	for k, v := range raw {
		if k == "customAttributes" || k == "_id" {
			continue
		}
		if knownItemFields[k] {
			known[k] = v
			continue
		}
		custom[k] = v
	}

	return known, custom
}

// validate checks required fields on item and, if idProyek is set, that it
// references an existing project. Returns a field->message map; empty means
// valid. This is the UC01/UC02 "validasiField" step.
func (h *ItemHandler) validate(ctx context.Context, item models.Item) map[string]string {
	fields := map[string]string{}

	required := map[string]string{
		"jenis":        item.Jenis,
		"serialNumber": item.SerialNumber,
		"nama":         item.Nama,
		"status":       item.Status,
		"idProyek":     item.IdProyek,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			fields[field] = field + " is required"
		}
	}

	if _, missing := fields["idProyek"]; !missing {
		exists, err := h.projectRepo.Exists(ctx, item.IdProyek)
		if err == nil && !exists {
			fields["idProyek"] = "referenced project does not exist"
		}
	}

	return fields
}

// CreateItem: POST /api/v1/items
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	raw := decodeBody(r)
	known, custom := splitCustomAttributes(raw)
	if len(custom) > 0 {
		known["customAttributes"] = custom
	}

	knownBytes, _ := json.Marshal(known)
	var item models.Item
	_ = json.Unmarshal(knownBytes, &item)

	if fields := h.validate(r.Context(), item); len(fields) > 0 {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid item fields", fields)
		return
	}

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
		"jenis", "serialNumber", "nama", "idProyek", "status",
		"licenseWindows", "licenseOffice",
		"credentials.account", "remoteInfo.ipAddress", "remoteInfo.anydesk", "remoteInfo.rustdesk",
		"q",
	} {
		if v := q.Get(field); v != "" {
			filters[field] = v
		}
	}
	// customAttributes keys are arbitrary and unknown ahead of time, so they
	// can't be enumerated in the fixed field list above — pass any
	// customAttributes.<key> query param straight through.
	for key, values := range q {
		if strings.HasPrefix(key, "customAttributes.") && len(values) > 0 && values[0] != "" {
			filters[key] = values[0]
		}
	}

	result, err := h.repo.List(r.Context(), repository.ListParams{
		Page:      page,
		Limit:     limit,
		SortBy:    q.Get("sortBy"),
		SortOrder: q.Get("sortOrder"),
		Filters:   filters,

		CreatedFrom: q.Get("createdFrom"),
		CreatedTo:   q.Get("createdTo"),
		UpdatedFrom: q.Get("updatedFrom"),
		UpdatedTo:   q.Get("updatedTo"),
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
	raw := decodeBody(r)
	known, custom := splitCustomAttributes(raw)

	existing, err := h.repo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Item not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update item", nil)
		return
	}

	merged := models.MergeItem(existing, known)
	if fields := h.validate(r.Context(), merged); len(fields) > 0 {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid item fields", fields)
		return
	}

	// customAttributes entries are $set by dotted key (customAttributes.foo),
	// not as a whole subdocument, so adding one ad-hoc attribute in this PATCH
	// doesn't wipe out attributes set by an earlier PATCH.
	patch := known
	for k, v := range custom {
		patch["customAttributes."+k] = v
	}

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
	data, err := h.repo.FilterOptions(r.Context(), "jenis", "status")
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

// FilterOptionsJenis: GET /api/v1/items/filter-options/jenis
func (h *ItemHandler) FilterOptionsJenis(w http.ResponseWriter, r *http.Request) {
	h.distinctFieldValues(w, r, "jenis")
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
	byJenis, err := h.repo.CountBy(ctx, "jenis")
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load stats", nil)
		return
	}
	byProyek, err := h.repo.CountBy(ctx, "idProyek")
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
		"totalItems":    total,
		"byStatus":      byStatus,
		"byJenis":       byJenis,
		"byProyek":      byProyek,
		"recentlyAdded": recentPublic,
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
