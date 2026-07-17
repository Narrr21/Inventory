package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"my-backend/internal/models"
	"my-backend/internal/repository"
	"my-backend/internal/response"
)

type ProjectHandler struct {
	repo *repository.ProjectRepository
}

func NewProjectHandler(repo *repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}

// CreateProject: POST /api/v1/projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	project := decodeProjectBody(r)

	if strings.TrimSpace(project.NamaProyek) == "" {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid project fields",
			map[string]string{"namaProyek": "namaProyek is required"})
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

func decodeProjectBody(r *http.Request) models.Project {
	var project models.Project
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&project)
	}
	return project
}
