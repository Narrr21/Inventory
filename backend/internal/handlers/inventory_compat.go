package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"my-backend/internal/models"
	"my-backend/internal/repository"
)

// InventoryCompatHandler serves a read-only, flat-shaped view of the real
// Item/Project data under /api/inventory — matching exactly what
// frontend/src/api/InventoryAPI.ts's fetchInventoryApi/fetchFilterOptionsApi
// already call (currently unreachable behind USE_MOCK=true), so flipping
// that one flag is enough to make the dashboard show real backend data.
// This does not replace /api/v1/items — that remains the documented
// api-contract.md surface; this is purely an additive compatibility view.
type InventoryCompatHandler struct {
	items    *repository.ItemRepository
	projects *repository.ProjectRepository
}

func NewInventoryCompatHandler(items *repository.ItemRepository, projects *repository.ProjectRepository) *InventoryCompatHandler {
	return &InventoryCompatHandler{items: items, projects: projects}
}

// compatItem mirrors frontend/src/types/dashboard.ts's BackendItem.
type compatItem struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	SerialNumber   string            `json:"serial_number"`
	LicenseWindows string            `json:"license_windows"`
	LicenseOffice  string            `json:"license_office"`
	Status         string            `json:"status"`
	Jenis          string            `json:"jenis"`
	Proyek         string            `json:"proyek"`
	CreatedAt      string            `json:"created_at,omitempty"`
	Credentials    map[string]string `json:"credentials,omitempty"`
	RemoteInfo     map[string]string `json:"remote_info,omitempty"`
	Other          map[string]string `json:"other,omitempty"`
}

type compatInventoryResponse struct {
	Rows  []compatItem `json:"rows"`
	Total int64        `json:"total"`
}

type compatFilterOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type compatFilterOptionsResponse struct {
	Project []compatFilterOption `json:"project"`
	Jenis   []compatFilterOption `json:"jenis"`
}

func writeCompatJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// splitNonEmpty splits a comma-joined query param (as sent by
// InventoryQueryParams.project.join(",") etc.) into its non-empty parts.
func splitNonEmpty(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func flattenCredentials(c models.Credentials) map[string]string {
	out := map[string]string{}
	if c.Account != "" {
		out["account"] = c.Account
	}
	if c.PasswordAccount != "" {
		out["passwordAccount"] = c.PasswordAccount
	}
	if c.PasswordPin != "" {
		out["passwordPin"] = c.PasswordPin
	}
	return out
}

func flattenRemoteInfo(r models.RemoteInfo) map[string]string {
	out := map[string]string{}
	if r.IPAddress != "" {
		out["ipAddress"] = r.IPAddress
	}
	if r.Anydesk != "" {
		out["anydesk"] = r.Anydesk
	}
	if r.Rustdesk != "" {
		out["rustdesk"] = r.Rustdesk
	}
	if r.PasswordRemote != "" {
		out["passwordRemote"] = r.PasswordRemote
	}
	return out
}

func stringifyCustomAttributes(m map[string]interface{}) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

func toCompatItem(item models.Item, projectNames map[string]string) compatItem {
	return compatItem{
		ID:             item.ID,
		Name:           item.Nama,
		SerialNumber:   item.SerialNumber,
		LicenseWindows: item.LicenseWindows,
		LicenseOffice:  item.LicenseOffice,
		Status:         item.Status,
		Jenis:          item.Jenis,
		Proyek:         projectNames[item.IdProyek],
		CreatedAt:      item.CreatedAt,
		Credentials:    flattenCredentials(item.Credentials),
		RemoteInfo:     flattenRemoteInfo(item.RemoteInfo),
		Other:          stringifyCustomAttributes(item.CustomAttributes),
	}
}

// projectNameMaps fetches every project once and returns id->namaProyek and
// namaProyek->id lookups, used both to resolve the compat endpoint's
// name-based project filter and to display a project name per item.
func (h *InventoryCompatHandler) projectNameMaps(ctx context.Context) (idToName map[string]string, nameToID map[string]string, err error) {
	projects, err := h.projects.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	idToName = make(map[string]string, len(projects))
	nameToID = make(map[string]string, len(projects))
	for _, p := range projects {
		idToName[p.ID] = p.NamaProyek
		nameToID[p.NamaProyek] = p.ID
	}
	return idToName, nameToID, nil
}

// List: GET /api/inventory
func (h *InventoryCompatHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ctx := r.Context()

	idToName, nameToID, err := h.projectNameMaps(ctx)
	if err != nil {
		http.Error(w, "failed to load projects", http.StatusInternalServerError)
		return
	}

	var projectIDs []string
	if names := splitNonEmpty(q.Get("project")); len(names) > 0 {
		for _, name := range names {
			if id, ok := nameToID[name]; ok {
				projectIDs = append(projectIDs, id)
			}
		}
		if len(projectIDs) == 0 {
			// None of the requested project names exist (anymore) — that's
			// zero results, not "no project filter at all".
			writeCompatJSON(w, http.StatusOK, compatInventoryResponse{Rows: []compatItem{}, Total: 0})
			return
		}
	}

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))

	result, err := h.items.ListCompat(ctx, repository.CompatListParams{
		Search:     q.Get("search"),
		Status:     q.Get("status"),
		ProjectIDs: projectIDs,
		Jenis:      splitNonEmpty(q.Get("jenis")),
		SortBy:     q.Get("sortKey"),
		SortOrder:  q.Get("sortDirection"),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		http.Error(w, "failed to list items", http.StatusInternalServerError)
		return
	}

	rows := make([]compatItem, 0, len(result.Items))
	for _, item := range result.Items {
		rows = append(rows, toCompatItem(item, idToName))
	}

	writeCompatJSON(w, http.StatusOK, compatInventoryResponse{Rows: rows, Total: result.Total})
}

// FilterOptions: GET /api/inventory/filter-options
func (h *InventoryCompatHandler) FilterOptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projects, err := h.projects.List(ctx)
	if err != nil {
		http.Error(w, "failed to load projects", http.StatusInternalServerError)
		return
	}
	jenisValues, err := h.items.Distinct(ctx, "jenis")
	if err != nil {
		http.Error(w, "failed to load jenis options", http.StatusInternalServerError)
		return
	}

	projectOptions := make([]compatFilterOption, 0, len(projects))
	for _, p := range projects {
		projectOptions = append(projectOptions, compatFilterOption{Value: p.NamaProyek, Label: p.NamaProyek})
	}
	jenisOptions := make([]compatFilterOption, 0, len(jenisValues))
	for _, j := range jenisValues {
		jenisOptions = append(jenisOptions, compatFilterOption{Value: j, Label: j})
	}

	writeCompatJSON(w, http.StatusOK, compatFilterOptionsResponse{Project: projectOptions, Jenis: jenisOptions})
}
