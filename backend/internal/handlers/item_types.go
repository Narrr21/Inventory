package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"my-backend/internal/models"
	"my-backend/internal/repository"
	"my-backend/internal/response"

	"github.com/go-chi/chi/v5"
)

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
// Blocked with 409 ITEM_TYPE_IN_USE while any item still carries this jenis
// (matched case-insensitively, same as the uniqueness rule — "Laptop" and
// "laptop" are one type, so items using either block deleting it).
//
// This replaces the earlier reassign-to-"Lainnya" behavior. Deleting a type
// now never touches an item document at all: rewriting the jenis of an
// unbounded number of items is a permanent, unauditable change (no soft
// delete, no audit log in this phase), and which type they should actually
// become is a call only the user can make. As a result "Lainnya" carries no
// special meaning anymore.
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

	itemCount, err := h.itemRepo.CountByJenis(ctx, itemType.Jenis)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete item type", nil)
		return
	}
	if itemCount > 0 {
		response.ErrDetails(w, http.StatusConflict, "ITEM_TYPE_IN_USE",
			"Item type is still used by "+strconv.FormatInt(itemCount, 10)+" item(s)",
			map[string]interface{}{"jenis": itemType.Jenis, "itemCount": itemCount})
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
		"_id":   id,
		"jenis": itemType.Jenis,
	}, nil)
}
