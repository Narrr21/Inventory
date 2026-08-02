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

type ItemTypeHandler struct {
	repo *repository.ItemTypeRepository
}

func NewItemTypeHandler(repo *repository.ItemTypeRepository) *ItemTypeHandler {
	return &ItemTypeHandler{repo: repo}
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
func (h *ItemTypeHandler) DeleteItemType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.repo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Item type not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete item type", nil)
		return
	}
	response.OK(w, http.StatusOK, map[string]string{"_id": id}, nil)
}
