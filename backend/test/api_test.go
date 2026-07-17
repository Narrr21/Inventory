// test/api_test.go
package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"my-backend/internal/db"
	"my-backend/internal/handlers"
	"my-backend/internal/repository"
)

// setupBackendAPI wires up the real backend (Mongo connection, repositories,
// handlers, chi router) exactly as cmd/server/main.go does, and serves it
// over a real HTTP listener so tests exercise the API as an outside client
// would. It connects to MONGODB_URI/MONGODB_TEST_DB (defaulting to a local
// instance and a dedicated test database), cleans the items/projects
// collections before and after the test, and skips if no MongoDB instance is
// reachable.
func setupBackendAPI(t *testing.T) string {
	t.Helper()

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://127.0.0.1:27017"
	}
	dbName := os.Getenv("MONGODB_TEST_DB")
	if dbName == "" {
		dbName = "rti_inventory_test"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := db.Connect(ctx, db.Config{URI: uri, Database: dbName})
	if err != nil {
		t.Skipf("skipping: no reachable MongoDB instance at %s: %v", uri, err)
	}

	itemsColl := client.Database.Collection("items")
	projectsColl := client.Database.Collection("projects")
	if _, err := itemsColl.DeleteMany(context.Background(), map[string]interface{}{}); err != nil {
		t.Fatalf("failed to clean items collection: %v", err)
	}
	if _, err := projectsColl.DeleteMany(context.Background(), map[string]interface{}{}); err != nil {
		t.Fatalf("failed to clean projects collection: %v", err)
	}

	itemRepo := repository.NewItemRepository(client.Database)
	projectRepo := repository.NewProjectRepository(client.Database)
	router := handlers.NewRouter(handlers.NewItemHandler(itemRepo, projectRepo), handlers.NewProjectHandler(projectRepo))
	server := httptest.NewServer(router)

	t.Cleanup(func() {
		server.Close()
		_, _ = itemsColl.DeleteMany(context.Background(), map[string]interface{}{})
		_, _ = projectsColl.DeleteMany(context.Background(), map[string]interface{}{})
		_ = db.Disconnect(context.Background(), client)
	})

	return server.URL
}

// apiRequest performs a real HTTP round trip against the running backend and
// decodes the standard success/error envelope.
func apiRequest(t *testing.T, baseURL, method, path string, body interface{}) (int, map[string]interface{}) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, baseURL+path, &buf)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var env map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return resp.StatusCode, env
}

// requestStatus performs a real HTTP round trip and returns only the status
// code, for routes that don't respond with the JSON envelope (e.g. chi's
// default 404 handler for unmatched routes).
func requestStatus(t *testing.T, baseURL, method, path string) int {
	t.Helper()

	req, err := http.NewRequest(method, baseURL+path, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode
}

func createProject(t *testing.T, baseURL string, overrides map[string]interface{}) map[string]interface{} {
	t.Helper()
	body := map[string]interface{}{
		"namaProyek": "ALPHA",
		"lokasi":     "Jakarta HQ",
	}
	for k, v := range overrides {
		body[k] = v
	}

	status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/projects", body)
	if status != http.StatusCreated {
		t.Fatalf("create project: status = %d, body = %+v", status, env)
	}
	return env["data"].(map[string]interface{})
}

func createItem(t *testing.T, baseURL string, overrides map[string]interface{}) map[string]interface{} {
	t.Helper()

	idProyek, ok := overrides["idProyek"].(string)
	if !ok || idProyek == "" {
		project := createProject(t, baseURL, nil)
		idProyek = project["_id"].(string)
	}

	body := map[string]interface{}{
		"jenis":        "Laptop",
		"serialNumber": "SN-TEST-001",
		"nama":         "RTI-TEST-001",
		"idProyek":     idProyek,
		"status":       "Active",
		"credentials": map[string]interface{}{
			"account": "user1", "passwordAccount": "pw1", "passwordPin": "1111",
		},
		"remoteInfo": map[string]interface{}{
			"ipAddress": "10.0.0.1", "anydesk": "111", "rustdesk": "222", "passwordRemote": "pw2",
		},
	}
	for k, v := range overrides {
		body[k] = v
	}

	status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/items", body)
	if status != http.StatusCreated {
		t.Fatalf("create item: status = %d, body = %+v", status, env)
	}
	return env["data"].(map[string]interface{})
}

// TestItemsLifecycle walks a single item through the full create -> read ->
// list -> update -> delete flow, as a real API consumer would.
func TestItemsLifecycle(t *testing.T) {
	baseURL := setupBackendAPI(t)

	var itemID string

	t.Run("Create", func(t *testing.T) {
		created := createItem(t, baseURL, map[string]interface{}{"nama": "RTI-LIFECYCLE-001"})
		itemID, _ = created["_id"].(string)
		if itemID == "" {
			t.Fatal("expected a generated _id")
		}
		if created["createdAt"] == "" || created["updatedAt"] == "" {
			t.Error("expected createdAt/updatedAt to be set")
		}
	})

	t.Run("Get", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/"+itemID, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["_id"] != itemID {
			t.Errorf("_id = %v, want %v", data["_id"], itemID)
		}
		credentials := data["credentials"].(map[string]interface{})
		if credentials["passwordPin"] != "1111" {
			t.Errorf("expected passwordPin visible on GetItem, got %+v", credentials)
		}
	})

	t.Run("List", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 1 {
			t.Fatalf("len(data) = %d, want 1", len(data))
		}
		meta := env["meta"].(map[string]interface{})
		if meta["total"].(float64) != 1 {
			t.Errorf("meta.total = %v, want 1", meta["total"])
		}

		first := data[0].(map[string]interface{})
		credentials := first["credentials"].(map[string]interface{})
		if _, ok := credentials["passwordAccount"]; ok {
			t.Error("expected credentials.passwordAccount to be omitted from list response")
		}
		if _, ok := credentials["passwordPin"]; ok {
			t.Error("expected credentials.passwordPin to be omitted from list response")
		}
		if credentials["account"] != "user1" {
			t.Errorf("expected credentials.account to survive redaction, got %+v", credentials)
		}
		remoteInfo := first["remoteInfo"].(map[string]interface{})
		if _, ok := remoteInfo["passwordRemote"]; ok {
			t.Error("expected remoteInfo.passwordRemote to be omitted from list response")
		}
		if remoteInfo["ipAddress"] != "10.0.0.1" || remoteInfo["anydesk"] != "111" || remoteInfo["rustdesk"] != "222" {
			t.Errorf("expected remoteInfo non-secret fields to survive redaction, got %+v", remoteInfo)
		}
	})

	t.Run("Update", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+itemID, map[string]interface{}{
			"status": "Maintenance",
		})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["status"] != "Maintenance" {
			t.Errorf("data = %+v, want status=Maintenance", data)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodDelete, "/api/v1/items/"+itemID, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
	})

	t.Run("GetAfterDelete", func(t *testing.T) {
		status, _ := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/"+itemID, nil)
		if status != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", status)
		}
	})
}

// TestItemNotFoundBehaviors checks that every mutating/reading route reports
// a consistent 404 envelope for an ID that doesn't exist.
func TestItemNotFoundBehaviors(t *testing.T) {
	baseURL := setupBackendAPI(t)
	const missingID = "000000000000000000000000"

	cases := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{"Get", http.MethodGet, "/api/v1/items/" + missingID, nil},
		{"Update", http.MethodPatch, "/api/v1/items/" + missingID, map[string]interface{}{"status": "Active"}},
		{"Delete", http.MethodDelete, "/api/v1/items/" + missingID, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, env := apiRequest(t, baseURL, tc.method, tc.path, tc.body)
			if status != http.StatusNotFound {
				t.Fatalf("status = %d, want 404, body = %+v", status, env)
			}
			if env["success"] != false {
				t.Errorf("success = %v, want false", env["success"])
			}
			errObj, ok := env["error"].(map[string]interface{})
			if !ok || errObj["code"] != "NOT_FOUND" {
				t.Errorf("error = %+v, want code NOT_FOUND", env["error"])
			}
		})
	}
}

// TestItemValidation exercises UC01/UC02's "value tidak valid" alternative
// flow: required fields must be present, and idProyek must reference a real
// project.
func TestItemValidation(t *testing.T) {
	baseURL := setupBackendAPI(t)

	assertValidationError := func(t *testing.T, env map[string]interface{}, wantFields ...string) {
		t.Helper()
		errObj, ok := env["error"].(map[string]interface{})
		if !ok || errObj["code"] != "VALIDATION_ERROR" {
			t.Fatalf("error = %+v, want code VALIDATION_ERROR", env["error"])
		}
		fields, ok := errObj["fields"].(map[string]interface{})
		if !ok {
			t.Fatalf("error.fields missing or wrong type: %+v", errObj)
		}
		for _, f := range wantFields {
			if _, ok := fields[f]; !ok {
				t.Errorf("expected error.fields to contain %q, got %+v", f, fields)
			}
		}
	}

	t.Run("CreateMissingRequiredFields", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/items", map[string]interface{}{
			"serialNumber": "SN-INVALID",
		})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422, body = %+v", status, env)
		}
		assertValidationError(t, env, "jenis", "nama", "status", "idProyek")
	})

	t.Run("CreateUnknownProject", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/items", map[string]interface{}{
			"jenis": "Laptop", "serialNumber": "SN-INVALID-2", "nama": "X", "status": "Active",
			"idProyek": "000000000000000000000000",
		})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422, body = %+v", status, env)
		}
		assertValidationError(t, env, "idProyek")
	})

	t.Run("UpdateClearingRequiredField", func(t *testing.T) {
		created := createItem(t, baseURL, nil)
		itemID := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+itemID, map[string]interface{}{
			"nama": "",
		})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422, body = %+v", status, env)
		}
		assertValidationError(t, env, "nama")
	})
}

// TestListEmptyResult checks UC03's "hasil kosong" alternative flow: a filter
// combination matching nothing returns an empty (not null) list and total 0.
func TestListEmptyResult(t *testing.T) {
	baseURL := setupBackendAPI(t)
	createItem(t, baseURL, map[string]interface{}{"jenis": "Laptop"})

	status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?jenis=NoSuchJenis", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, body = %+v", status, env)
	}
	data := env["data"].([]interface{})
	if len(data) != 0 {
		t.Errorf("data = %v, want empty", data)
	}
	meta := env["meta"].(map[string]interface{})
	if meta["total"].(float64) != 0 {
		t.Errorf("meta.total = %v, want 0", meta["total"])
	}
}

// TestListFilterSearchSortPagination exercises F02/F03/F07/F08: equality
// filters, free-text search, sort direction, pagination, and the createdAt
// range filter.
func TestListFilterSearchSortPagination(t *testing.T) {
	baseURL := setupBackendAPI(t)

	projectA := createProject(t, baseURL, map[string]interface{}{"namaProyek": "ALPHA"})
	projectB := createProject(t, baseURL, map[string]interface{}{"namaProyek": "BETA"})

	createItem(t, baseURL, map[string]interface{}{
		"nama": "RTI-AAA", "jenis": "Laptop", "status": "Healthy", "idProyek": projectA["_id"],
	})
	createItem(t, baseURL, map[string]interface{}{
		"nama": "RTI-BBB", "jenis": "Monitor", "status": "Broken", "idProyek": projectB["_id"],
	})
	createItem(t, baseURL, map[string]interface{}{
		"nama": "RTI-CCC", "jenis": "Laptop", "status": "Healthy", "idProyek": projectA["_id"],
	})

	t.Run("FilterByJenis", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?jenis=Monitor", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 1 {
			t.Fatalf("len(data) = %d, want 1", len(data))
		}
	})

	t.Run("FilterByIdProyek", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?idProyek="+projectA["_id"].(string), nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 2 {
			t.Fatalf("len(data) = %d, want 2", len(data))
		}
	})

	t.Run("Search", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?q=BBB", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 1 {
			t.Fatalf("len(data) = %d, want 1", len(data))
		}
	})

	t.Run("SortAscDesc", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?sortBy=nama&sortOrder=asc", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		first := data[0].(map[string]interface{})
		if first["nama"] != "RTI-AAA" {
			t.Errorf("first.nama = %v, want RTI-AAA (asc)", first["nama"])
		}

		status, env = apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?sortBy=nama&sortOrder=desc", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data = env["data"].([]interface{})
		first = data[0].(map[string]interface{})
		if first["nama"] != "RTI-CCC" {
			t.Errorf("first.nama = %v, want RTI-CCC (desc)", first["nama"])
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?sortBy=nama&sortOrder=asc&page=2&limit=1", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 1 {
			t.Fatalf("len(data) = %d, want 1", len(data))
		}
		first := data[0].(map[string]interface{})
		if first["nama"] != "RTI-BBB" {
			t.Errorf("page 2 item.nama = %v, want RTI-BBB", first["nama"])
		}
		meta := env["meta"].(map[string]interface{})
		if meta["totalPages"].(float64) != 3 {
			t.Errorf("meta.totalPages = %v, want 3", meta["totalPages"])
		}
	})

	t.Run("CreatedAtRange", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?createdFrom=9999-01-01T00:00:00.000Z", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		if data := env["data"].([]interface{}); len(data) != 0 {
			t.Errorf("createdFrom in the far future: len(data) = %d, want 0", len(data))
		}

		status, env = apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?createdFrom=2000-01-01T00:00:00.000Z", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		if data := env["data"].([]interface{}); len(data) != 3 {
			t.Errorf("createdFrom in the past: len(data) = %d, want 3", len(data))
		}
	})
}

// TestFilterOptionsAndStats exercises the aggregate endpoints against seeded data.
func TestFilterOptionsAndStats(t *testing.T) {
	baseURL := setupBackendAPI(t)

	createItem(t, baseURL, map[string]interface{}{
		"nama": "RTI-AGG-001", "jenis": "Laptop", "status": "Active",
	})
	createItem(t, baseURL, map[string]interface{}{
		"nama": "RTI-AGG-002", "jenis": "Monitor", "status": "Idle",
	})

	t.Run("FilterOptions", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/filter-options", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		jenis := data["jenis"].([]interface{})
		if len(jenis) != 2 {
			t.Errorf("jenis = %v, want 2 distinct values", jenis)
		}
	})

	t.Run("FilterOptionsJenis", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/filter-options/jenis", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 2 {
			t.Errorf("data = %v, want 2 distinct values", data)
		}
	})

	t.Run("Stats", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/stats", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["totalItems"].(float64) != 2 {
			t.Errorf("totalItems = %v, want 2", data["totalItems"])
		}
		byStatus := data["byStatus"].([]interface{})
		if len(byStatus) != 2 {
			t.Errorf("byStatus buckets = %d, want 2", len(byStatus))
		}
		byJenis := data["byJenis"].([]interface{})
		if len(byJenis) != 2 {
			t.Errorf("byJenis buckets = %d, want 2", len(byJenis))
		}
		recent := data["recentlyAdded"].([]interface{})
		if len(recent) != 2 {
			t.Errorf("recentlyAdded = %d, want 2", len(recent))
		}
	})
}

// TestProjectsCreateAndList exercises the minimal Proyek reference endpoints.
func TestProjectsCreateAndList(t *testing.T) {
	baseURL := setupBackendAPI(t)

	created := createProject(t, baseURL, map[string]interface{}{
		"namaProyek": "DELTA", "lokasi": "Bandung Office",
	})
	if created["_id"] == "" {
		t.Fatal("expected a generated _id")
	}

	t.Run("List", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/projects", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 1 {
			t.Fatalf("len(data) = %d, want 1", len(data))
		}
		first := data[0].(map[string]interface{})
		if first["namaProyek"] != "DELTA" || first["lokasi"] != "Bandung Office" {
			t.Errorf("data[0] = %+v, want namaProyek=DELTA lokasi=Bandung Office", first)
		}
	})

	t.Run("CreateMissingNamaProyek", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/projects", map[string]interface{}{
			"lokasi": "Somewhere",
		})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422, body = %+v", status, env)
		}
		errObj := env["error"].(map[string]interface{})
		if errObj["code"] != "VALIDATION_ERROR" {
			t.Errorf("error.code = %v, want VALIDATION_ERROR", errObj["code"])
		}
	})
}

// TestCustomAttributesAutoRouting exercises the fix for unrecognized fields:
// they must never be silently dropped (as Create used to) or written straight
// into the document root ungoverned (as Update used to) — they always land in
// customAttributes, and successive PATCHes must not clobber each other's
// ad-hoc attributes.
func TestCustomAttributesAutoRouting(t *testing.T) {
	baseURL := setupBackendAPI(t)

	t.Run("CreateRoutesUnknownFieldToCustomAttributes", func(t *testing.T) {
		created := createItem(t, baseURL, map[string]interface{}{
			"garansiTahun": float64(3),
		})
		custom := created["customAttributes"].(map[string]interface{})
		if custom["garansiTahun"] != float64(3) {
			t.Errorf("customAttributes = %+v, want garansiTahun=3", custom)
		}
		if _, ok := created["garansiTahun"]; ok {
			t.Error("expected garansiTahun to NOT appear at the top level")
		}
	})

	t.Run("SuccessivePatchesPreserveEarlierCustomAttributes", func(t *testing.T) {
		created := createItem(t, baseURL, nil)
		itemID := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+itemID, map[string]interface{}{
			"warna": "Merah",
		})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}

		status, env = apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+itemID, map[string]interface{}{
			"beratKg": float64(2),
		})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}

		status, env = apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/"+itemID, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		custom := data["customAttributes"].(map[string]interface{})
		if custom["warna"] != "Merah" {
			t.Errorf("expected warna to survive the second patch, customAttributes = %+v", custom)
		}
		if custom["beratKg"] != float64(2) {
			t.Errorf("expected beratKg to be set, customAttributes = %+v", custom)
		}
	})

}

// TestFilterAndSortByCustomAttribute checks that customAttributes.<key> works
// as a dot-notation filter/sort param, closing the gap where the sanctioned
// ad-hoc-column mechanism used to be unqueryable. Runs in its own clean
// collection since sort-by-missing-field semantics would be muddied by items
// left over from other tests.
func TestFilterAndSortByCustomAttribute(t *testing.T) {
	baseURL := setupBackendAPI(t)

	createItem(t, baseURL, map[string]interface{}{"nama": "RTI-CA-A", "warna": "Merah"})
	createItem(t, baseURL, map[string]interface{}{"nama": "RTI-CA-B", "warna": "Biru"})

	t.Run("Filter", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?customAttributes.warna=Biru", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 1 {
			t.Fatalf("len(data) = %d, want 1", len(data))
		}
		first := data[0].(map[string]interface{})
		if first["nama"] != "RTI-CA-B" {
			t.Errorf("nama = %v, want RTI-CA-B", first["nama"])
		}
	})

	t.Run("Sort", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?sortBy=customAttributes.warna&sortOrder=asc", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		first := data[0].(map[string]interface{})
		if first["nama"] != "RTI-CA-B" {
			t.Errorf("first.nama = %v, want RTI-CA-B (Biru sorts before Merah)", first["nama"])
		}
	})
}

// TestUnknownRouteReturns404 checks routing behavior for the API as a whole,
// independent of any specific resource.
func TestUnknownRouteReturns404(t *testing.T) {
	baseURL := setupBackendAPI(t)

	status := requestStatus(t, baseURL, http.MethodGet, "/api/v1/does-not-exist")
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
}
