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

type ProjectHandler struct {
	repo     *repository.ProjectRepository
	itemRepo *repository.ItemRepository
}

func NewProjectHandler(repo *repository.ProjectRepository, itemRepo *repository.ItemRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo, itemRepo: itemRepo}
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

// coordNumber reads one coordinate component. A JSON number is the expected
// form, but a numeric string is accepted too: HTML number/text inputs hand
// back strings, and rejecting "-6.2088" would push that conversion onto every
// client for no benefit. Anything else (bool, object, "abc", "") is a real
// error, not something to coerce to 0 — 0,0 is a legitimate point in the
// Atlantic, so a silent fallback would plant a pin there.
func coordNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

// parseKoordinat interprets the "koordinat" key of a request body, returning
// the parsed point (nil = no point) and per-field validation errors.
//
// Three distinct inputs, three distinct meanings:
//   - key absent          -> present=false, caller leaves the existing value alone
//   - explicit null       -> present=true, koordinat=nil, meaning "remove the point"
//   - object              -> must carry BOTH lat and lng, in range
//
// An empty object is deliberately NOT treated as "remove": a form that
// serializes {lat: undefined, lng: undefined} would otherwise silently wipe a
// correct point. It fails loudly as a missing-field error instead.
func parseKoordinat(raw map[string]interface{}) (koordinat *models.Koordinat, present bool, fields map[string]string) {
	value, present := raw["koordinat"]
	if !present {
		return nil, false, nil
	}
	if value == nil {
		return nil, true, nil
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return nil, true, map[string]string{"koordinat": "koordinat must be an object with lat and lng"}
	}

	fields = map[string]string{}
	var lat, lng float64

	for _, c := range []struct {
		key    string
		limit  float64
		target *float64
	}{
		{"lat", 90, &lat},
		{"lng", 180, &lng},
	} {
		rawValue, exists := obj[c.key]
		if !exists || rawValue == nil {
			fields["koordinat."+c.key] = c.key + " is required"
			continue
		}
		parsed, ok := coordNumber(rawValue)
		if !ok {
			fields["koordinat."+c.key] = c.key + " must be a number"
			continue
		}
		if parsed < -c.limit || parsed > c.limit {
			fields["koordinat."+c.key] = c.key + " must be between -" +
				strconv.FormatFloat(c.limit, 'f', -1, 64) + " and " +
				strconv.FormatFloat(c.limit, 'f', -1, 64)
			continue
		}
		*c.target = parsed
	}

	if len(fields) > 0 {
		return nil, true, fields
	}
	return &models.Koordinat{Lat: lat, Lng: lng}, true, nil
}

// CreateProject: POST /api/v1/projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	raw, err := decodeJSONObject(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "MALFORMED_BODY", "Request body is not a valid JSON object", nil)
		return
	}
	project := projectFromRaw(raw)

	koordinat, _, fields := parseKoordinat(raw)
	if fields == nil {
		fields = map[string]string{}
	}
	project.Koordinat = koordinat

	if strings.TrimSpace(project.NamaProyek) == "" {
		fields["namaProyek"] = "namaProyek is required"
	} else {
		exists, err := h.repo.ExistsByNamaProyek(r.Context(), project.NamaProyek, "")
		if err != nil {
			response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create project", nil)
			return
		}
		if exists {
			fields["namaProyek"] = "namaProyek already exists"
		}
	}

	// Every problem with the body is reported in one response — a form that
	// got both the name and the coordinates wrong shouldn't need two attempts
	// to find that out.
	if len(fields) > 0 {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid project fields", fields)
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
	// koordinat is a whole-object replace, like credentials/remoteInfo on
	// Item: an omitted key keeps the existing point, an explicit null removes
	// it, and an object must be complete. There's no per-component patch.
	koordinat, koordinatSent, fields := parseKoordinat(raw)
	if koordinatSent {
		merged.Koordinat = koordinat
	}
	if fields == nil {
		fields = map[string]string{}
	}

	if strings.TrimSpace(merged.NamaProyek) == "" {
		fields["namaProyek"] = "namaProyek is required"
	} else {
		exists, err := h.repo.ExistsByNamaProyek(r.Context(), merged.NamaProyek, id)
		if err != nil {
			response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update project", nil)
			return
		}
		if exists {
			fields["namaProyek"] = "namaProyek already exists"
		}
	}

	if len(fields) > 0 {
		response.Err(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid project fields", fields)
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
//
// Blocked with 409 PROJECT_IN_USE while any item still references the project.
// Neither alternative is acceptable here: cascade-deleting the items would
// destroy inventory records to remove a label, and letting them keep a dangling
// idProyek (the pre-v6 behavior) left items whose project name silently
// vanished from every screen with no way to find or fix them. So the decision
// is handed back to the user, with the blocking count included so the UI can
// point them at the items.
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	project, err := h.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "Project not found", nil)
		return
	}
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete project", nil)
		return
	}

	itemCount, err := h.itemRepo.CountByProject(ctx, id)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete project", nil)
		return
	}
	if itemCount > 0 {
		response.ErrDetails(w, http.StatusConflict, "PROJECT_IN_USE",
			"Project still has "+strconv.FormatInt(itemCount, 10)+" item(s)",
			map[string]interface{}{"namaProyek": project.NamaProyek, "itemCount": itemCount})
		return
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Err(w, http.StatusNotFound, "NOT_FOUND", "Project not found", nil)
			return
		}
		response.Err(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete project", nil)
		return
	}
	response.OK(w, http.StatusOK, map[string]string{"_id": id, "namaProyek": project.NamaProyek}, nil)
}
