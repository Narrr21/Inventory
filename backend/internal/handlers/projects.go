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

type ProjectHandler struct {
	repo *repository.ProjectRepository
}

func NewProjectHandler(repo *repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}


func projectFromRaw(raw map[string]interface{}) models.Project {
	var project models.Project
	if v, ok := raw["namaProyek"].(string); ok {
		project.NamaProyek = v
	}
	if v, ok := raw["lokasi"].(string); ok {
		project.Lokasi = v
	}
	return project
}

// CreateProject: POST /api/v1/projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	raw, err := decodeJSONObject(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "MALFORMED_BODY", "Request body is not a valid JSON object", nil)
		return
	}
	project := projectFromRaw(raw)

	if strings.TrimSpace(project.NamaProyek) == "" {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid project fields",
			map[string]string{"namaProyek": "namaProyek is required"})
		return
	}

	exists, err := h.repo.ExistsByNamaProyek(r.Context(), project.NamaProyek, "")
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create project", nil)
		return
	}
	if exists {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid project fields",
			map[string]string{"namaProyek": "namaProyek already exists"})
		return
	}

	created, err := h.repo.Create(r.Context(), project)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create project", nil)
		return
	}
	response.OK(w, http.StatusCreated, created, nil)
}

// ListProjects: GET /api/v1/projects
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.repo.List(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list projects", nil)
		return
	}
	response.OK(w, http.StatusOK, projects, nil)
}

// GetProject: GET /api/v1/projects/{id}
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	project, err := h.repo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Project not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get project", nil)
		return
	}
	response.OK(w, http.StatusOK, project, nil)
}

// UpdateProject: PATCH /api/v1/projects/{id}
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	raw, err := decodeJSONObject(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "MALFORMED_BODY", "Request body is not a valid JSON object", nil)
		return
	}

	existing, err := h.repo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Project not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update project", nil)
		return
	}

	merged := existing
	if v, ok := raw["namaProyek"]; ok {
		if s, ok := v.(string); ok {
			merged.NamaProyek = s
		}
	}
	if v, ok := raw["lokasi"]; ok {
		if s, ok := v.(string); ok {
			merged.Lokasi = s
		}
	}

	if strings.TrimSpace(merged.NamaProyek) == "" {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid project fields",
			map[string]string{"namaProyek": "namaProyek is required"})
		return
	}

	exists, err := h.repo.ExistsByNamaProyek(r.Context(), merged.NamaProyek, id)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update project", nil)
		return
	}
	if exists {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid project fields",
			map[string]string{"namaProyek": "namaProyek already exists"})
		return
	}

	updated, err := h.repo.Update(r.Context(), id, merged)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Project not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update project", nil)
		return
	}
	response.OK(w, http.StatusOK, updated, nil)
}

// DeleteProject: DELETE /api/v1/projects/{id}
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.repo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Project not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete project", nil)
		return
	}
	response.OK(w, http.StatusOK, map[string]string{"_id": id}, nil)
}
