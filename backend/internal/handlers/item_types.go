package handlers

import (
	"errors"
	"net/http"
	"strings"

	"my-backend/internal/models"
	"my-backend/internal/repository"
	"my-backend/internal/response"

	"github.com/go-chi/chi/v5"
)

// DefaultJenis is the fallback item type every item falls back to when the
// type it was using gets deleted. It is auto-created in the master list on
// demand, and cannot itself be deleted — otherwise a delete would have
// nothing left to fall back to.
const DefaultJenis = "Lainnya"

type ItemTypeHandler struct {
	repo     *repository.ItemTypeRepository
	itemRepo *repository.ItemRepository
}

func NewItemTypeHandler(repo *repository.ItemTypeRepository, itemRepo *repository.ItemRepository) *ItemTypeHandler {
	return &ItemTypeHandler{repo: repo, itemRepo: itemRepo}
}

// CreateItemType: POST /api/v1/item-types
func (h *ItemTypeHandler) CreateItemType(w http.ResponseWriter, r *http.Request) {
	raw, err := decodeJSONObject(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "MALFORMED_BODY", "Request body is not a valid JSON object", nil)
		return
	}

	jenis, _ := raw["jenis"].(string)
	jenis = strings.TrimSpace(jenis)
	if jenis == "" {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid item type fields",
			map[string]string{"jenis": "jenis is required"})
		return
	}

	exists, err := h.repo.ExistsByJenis(r.Context(), jenis)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create item type", nil)
		return
	}
	if exists {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid item type fields",
			map[string]string{"jenis": "jenis already exists"})
		return
	}

	created, err := h.repo.Create(r.Context(), models.ItemType{Jenis: jenis})
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create item type", nil)
		return
	}
	response.OK(w, http.StatusCreated, created, nil)
}

// ListItemTypes: GET /api/v1/item-types
func (h *ItemTypeHandler) ListItemTypes(w http.ResponseWriter, r *http.Request) {
	itemTypes, err := h.repo.List(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list item types", nil)
		return
	}
	response.OK(w, http.StatusOK, itemTypes, nil)
}

// DeleteItemType: DELETE /api/v1/item-types/{id}
//
// Deleting a type is never blocked by items still using it — instead those
// items are reassigned to DefaultJenis first, so no item is ever left with a
// jenis that has no entry in the master list. That reassignment is a real,
// permanent rewrite of those item documents (no soft delete, no audit log in
// this phase), which is why the response reports how many were touched.
func (h *ItemTypeHandler) DeleteItemType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	itemType, err := h.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Item type not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete item type", nil)
		return
	}

	if strings.EqualFold(strings.TrimSpace(itemType.Jenis), DefaultJenis) {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid item type fields",
			map[string]string{"jenis": "the default item type cannot be deleted"})
		return
	}

	// Guarantee the fallback exists before pointing items at it, so the
	// invariant holds even on a database that has never seen it before.
	if err := h.repo.EnsureExists(ctx, DefaultJenis); err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete item type", nil)
		return
	}

	reassigned, err := h.itemRepo.ReassignJenis(ctx, itemType.Jenis, DefaultJenis)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete item type", nil)
		return
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Err(w, http.StatusNotFound, "NOT_FOUND", "Item type not found", nil)
			return
		}
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete item type", nil)
		return
	}

	response.OK(w, http.StatusOK, map[string]interface{}{
		"_id":             id,
		"jenis":           itemType.Jenis,
		"defaultJenis":    DefaultJenis,
		"reassignedItems": reassigned,
	}, nil)
}
