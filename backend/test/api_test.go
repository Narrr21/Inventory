// test/api_test.go
package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"my-backend/internal/db"
	"my-backend/internal/handlers"
	"my-backend/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
	baseURL, _ := setupBackendAPIWithDB(t)
	return baseURL
}

// setupBackendAPIWithDB is setupBackendAPI plus the raw database handle, for
// the few tests that have to manufacture states the API itself no longer
// allows — e.g. an item whose idProyek points nowhere, which used to be
// reachable by deleting a project out from under it and is now blocked.
func setupBackendAPIWithDB(t *testing.T) (string, *mongo.Database) {
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
	itemTypesColl := client.Database.Collection("itemTypes")
	if _, err := itemsColl.DeleteMany(context.Background(), map[string]interface{}{}); err != nil {
		t.Fatalf("failed to clean items collection: %v", err)
	}
	if _, err := projectsColl.DeleteMany(context.Background(), map[string]interface{}{}); err != nil {
		t.Fatalf("failed to clean projects collection: %v", err)
	}
	if _, err := itemTypesColl.DeleteMany(context.Background(), map[string]interface{}{}); err != nil {
		t.Fatalf("failed to clean itemTypes collection: %v", err)
	}

	itemRepo := repository.NewItemRepository(client.Database)
	projectRepo := repository.NewProjectRepository(client.Database)
	itemTypeRepo := repository.NewItemTypeRepository(client.Database)
	router := handlers.NewRouter(
		handlers.NewItemHandler(itemRepo, projectRepo, itemTypeRepo),
		handlers.NewProjectHandler(projectRepo, itemRepo),
		handlers.NewItemTypeHandler(itemTypeRepo, itemRepo),
		handlers.NewAnalyticsHandler(itemRepo, projectRepo, itemTypeRepo),
	)
	server := httptest.NewServer(router)

	t.Cleanup(func() {
		server.Close()
		_, _ = itemsColl.DeleteMany(context.Background(), map[string]interface{}{})
		_, _ = projectsColl.DeleteMany(context.Background(), map[string]interface{}{})
		_, _ = itemTypesColl.DeleteMany(context.Background(), map[string]interface{}{})
		_ = db.Disconnect(context.Background(), client)
	})

	return server.URL, client.Database
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

// assertErrorEnvelope checks the flat failure envelope every endpoint shares:
// success=false, a numeric code mirroring the HTTP status, the expected
// errorCode, and a non-empty msg. Returns the envelope so callers can go on to
// inspect fields/details.
func assertErrorEnvelope(t *testing.T, status int, env map[string]interface{}, wantStatus int, wantErrorCode string) map[string]interface{} {
	t.Helper()
	if status != wantStatus {
		t.Fatalf("status = %d, want %d, body = %+v", status, wantStatus, env)
	}
	if env["success"] != false {
		t.Errorf("success = %v, want false", env["success"])
	}
	if code, ok := env["code"].(float64); !ok || int(code) != wantStatus {
		t.Errorf("code = %v, want %d (must mirror the HTTP status)", env["code"], wantStatus)
	}
	if env["errorCode"] != wantErrorCode {
		t.Errorf("errorCode = %v, want %v", env["errorCode"], wantErrorCode)
	}
	if msg, ok := env["msg"].(string); !ok || msg == "" {
		t.Errorf("msg = %v, want a non-empty explanation", env["msg"])
	}
	if _, nested := env["error"]; nested {
		t.Errorf("envelope still carries the pre-v6 nested error object: %+v", env["error"])
	}
	return env
}

// assertErrorFields asserts an error envelope carries per-field validation
// messages for every named field.
func assertErrorFields(t *testing.T, env map[string]interface{}, wantFields ...string) {
	t.Helper()
	fields, ok := env["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("fields missing or wrong type: %+v", env)
	}
	for _, f := range wantFields {
		if _, ok := fields[f]; !ok {
			t.Errorf("expected fields to contain %q, got %+v", f, fields)
		}
	}
}

// assertSuccessCode checks a success envelope carries the numeric code too —
// the whole point of adding it was that clients can read one field regardless
// of outcome.
func assertSuccessCode(t *testing.T, env map[string]interface{}, wantStatus int) {
	t.Helper()
	if env["success"] != true {
		t.Errorf("success = %v, want true", env["success"])
	}
	if code, ok := env["code"].(float64); !ok || int(code) != wantStatus {
		t.Errorf("code = %v, want %d", env["code"], wantStatus)
	}
}

// projectSeq gives each auto-created default project a distinct namaProyek,
// since that field is now uniqueness-validated server-side and several
// tests/helpers call createProject(t, baseURL, nil) more than once against
// the same shared per-test database.
var projectSeq int

func createProject(t *testing.T, baseURL string, overrides map[string]interface{}) map[string]interface{} {
	t.Helper()
	projectSeq++
	body := map[string]interface{}{
		"namaProyek": fmt.Sprintf("ALPHA-%d", projectSeq),
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

func createItemType(t *testing.T, baseURL string, jenis string) map[string]interface{} {
	t.Helper()
	status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/item-types", map[string]interface{}{"jenis": jenis})
	if status != http.StatusCreated {
		t.Fatalf("create item type: status = %d, body = %+v", status, env)
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

		// credentials/remoteInfo are free-form, like customAttributes — no
		// server-side redaction, so the list response carries the same
		// content as GetItem.
		first := data[0].(map[string]interface{})
		credentials := first["credentials"].(map[string]interface{})
		if credentials["passwordAccount"] != "pw1" {
			t.Errorf("expected credentials.passwordAccount in list response, got %+v", credentials)
		}
		if credentials["passwordPin"] != "1111" {
			t.Errorf("expected credentials.passwordPin in list response, got %+v", credentials)
		}
		if credentials["account"] != "user1" {
			t.Errorf("expected credentials.account in list response, got %+v", credentials)
		}
		remoteInfo := first["remoteInfo"].(map[string]interface{})
		if remoteInfo["passwordRemote"] != "pw2" {
			t.Errorf("expected remoteInfo.passwordRemote in list response, got %+v", remoteInfo)
		}
		if remoteInfo["ipAddress"] != "10.0.0.1" || remoteInfo["anydesk"] != "111" || remoteInfo["rustdesk"] != "222" {
			t.Errorf("expected remoteInfo fields in list response, got %+v", remoteInfo)
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
			assertErrorEnvelope(t, status, env, http.StatusNotFound, "NOT_FOUND")
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
		if env["errorCode"] != "VALIDATION_ERROR" {
			t.Fatalf("errorCode = %v, want VALIDATION_ERROR (body = %+v)", env["errorCode"], env)
		}
		assertErrorFields(t, env, wantFields...)
	}

	t.Run("CreateMissingRequiredFields", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/items", map[string]interface{}{
			"serialNumber": "SN-INVALID",
		})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422, body = %+v", status, env)
		}
		assertValidationError(t, env, "jenis", "status", "idProyek")
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
			"status": "",
		})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422, body = %+v", status, env)
		}
		assertValidationError(t, env, "status")
	})

	// nama is explicitly optional per CLAUDE.md/api-contract.md — create and
	// update must both succeed without it.
	t.Run("CreateWithoutNamaSucceeds", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/items", map[string]interface{}{
			"jenis": "Laptop", "serialNumber": "SN-NO-NAMA", "status": "Active",
			"idProyek": createProject(t, baseURL, nil)["_id"],
		})
		if status != http.StatusCreated {
			t.Fatalf("status = %d, want 201, body = %+v", status, env)
		}
	})

	t.Run("UpdateClearingNamaSucceeds", func(t *testing.T) {
		created := createItem(t, baseURL, nil)
		itemID := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+itemID, map[string]interface{}{
			"nama": "",
		})
		if status != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["nama"] != "" {
			t.Errorf("nama = %v, want empty string", data["nama"])
		}
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

	t.Run("FilterByNamaProyek", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?namaProyek="+projectA["namaProyek"].(string), nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 2 {
			t.Fatalf("len(data) = %d, want 2", len(data))
		}
	})

	t.Run("FilterByNamaProyekUnknownReturnsEmpty", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?namaProyek=NoSuchProject", nil)
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
	})

	t.Run("FilterByIdProyekTakesPrecedenceOverNamaProyek", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet,
			"/api/v1/items?idProyek="+projectB["_id"].(string)+"&namaProyek="+projectA["namaProyek"].(string), nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) != 1 {
			t.Fatalf("len(data) = %d, want 1 (projectB's single item)", len(data))
		}
		first := data[0].(map[string]interface{})
		if first["idProyek"] != projectB["_id"] {
			t.Errorf("idProyek = %v, want %v (literal idProyek should win)", first["idProyek"], projectB["_id"])
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

		// project must carry full Project objects (_id + namaProyek + lokasi)
		// so a client never needs a separate GET /projects call just to
		// populate an idProyek/namaProyek dropdown.
		projects := data["project"].([]interface{})
		if len(projects) != 2 {
			t.Fatalf("project = %v, want 2 entries (one per auto-created project)", projects)
		}
		first := projects[0].(map[string]interface{})
		for _, field := range []string{"_id", "namaProyek", "lokasi"} {
			if _, ok := first[field]; !ok {
				t.Errorf("project entry missing %q: %+v", field, first)
			}
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

	// jenis comes from the item-type master list, not from distinct values
	// on items — so a type registered ahead of any item that uses it still
	// shows up as a dropdown option.
	t.Run("FilterOptionsJenisIncludesUnusedType", func(t *testing.T) {
		createItemType(t, baseURL, "Proyektor")

		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/filter-options", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		found := false
		for _, raw := range env["data"].(map[string]interface{})["jenis"].([]interface{}) {
			if raw.(string) == "Proyektor" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected Proyektor (registered but unused) in jenis options, got %+v", env["data"])
		}
	})

	// The F05 stub is gone; /analytics/summary replaced it.
	t.Run("StatsEndpointRemoved", func(t *testing.T) {
		if status := requestStatus(t, baseURL, http.MethodGet, "/api/v1/items/stats"); status != http.StatusNotFound {
			t.Errorf("GET /items/stats status = %d, want 404 (replaced by /analytics/summary)", status)
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
		assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
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

// TestCredentialsAndRemoteInfoAreFreeForm exercises the fix for a bug where
// Credentials/RemoteInfo were decoded into fixed-shape Go structs, so any
// sub-key outside the documented set (account/passwordAccount/passwordPin,
// ipAddress/anydesk/rustdesk/passwordRemote) was silently dropped instead of
// stored — a client sending an arbitrary key got back a 201 with an empty
// object and no indication anything was lost. Like customAttributes, these
// are schemaless: whatever key/value pairs are sent must round-trip as-is.
func TestCredentialsAndRemoteInfoAreFreeForm(t *testing.T) {
	baseURL := setupBackendAPI(t)

	t.Run("CreatePreservesArbitraryKeys", func(t *testing.T) {
		created := createItem(t, baseURL, map[string]interface{}{
			"credentials": map[string]interface{}{"111": "12"},
			"remoteInfo":  map[string]interface{}{"122": "13"},
		})

		credentials := created["credentials"].(map[string]interface{})
		if credentials["111"] != "12" {
			t.Errorf("credentials = %+v, want 111=12 preserved", credentials)
		}
		remoteInfo := created["remoteInfo"].(map[string]interface{})
		if remoteInfo["122"] != "13" {
			t.Errorf("remoteInfo = %+v, want 122=13 preserved", remoteInfo)
		}
	})

	t.Run("UnsetCredentialsDefaultToEmptyObjectNotNull", func(t *testing.T) {
		created := createItem(t, baseURL, map[string]interface{}{
			"credentials": map[string]interface{}{},
			"remoteInfo":  map[string]interface{}{},
		})
		if _, ok := created["credentials"].(map[string]interface{}); !ok {
			t.Errorf("credentials = %+v (%T), want {} not null", created["credentials"], created["credentials"])
		}
		if _, ok := created["remoteInfo"].(map[string]interface{}); !ok {
			t.Errorf("remoteInfo = %+v (%T), want {} not null", created["remoteInfo"], created["remoteInfo"])
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

// TestProjectsFullCRUD exercises UC06-UC08 (Get/Update/Delete Proyek), the
// endpoints the v4 api-contract added on top of the create+list pair that
// used to be all Project supported.
func TestProjectsFullCRUD(t *testing.T) {
	baseURL := setupBackendAPI(t)

	t.Run("Get", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{"namaProyek": "EPSILON"})
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/projects/"+id, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["namaProyek"] != "EPSILON" {
			t.Errorf("namaProyek = %v, want EPSILON", data["namaProyek"])
		}
	})

	t.Run("GetNotFound", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/projects/000000000000000000000000", nil)
		if status != http.StatusNotFound {
			t.Fatalf("status = %d, want 404, body = %+v", status, env)
		}
	})

	t.Run("Update", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{"namaProyek": "ZETA", "lokasi": "Jakarta HQ"})
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/"+id, map[string]interface{}{
			"lokasi": "Bandung Office - Lantai 2",
		})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["namaProyek"] != "ZETA" || data["lokasi"] != "Bandung Office - Lantai 2" {
			t.Errorf("data = %+v, want namaProyek=ZETA lokasi=Bandung Office - Lantai 2", data)
		}
	})

	t.Run("UpdateBlankNamaProyekRejected", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{"namaProyek": "ETA"})
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/"+id, map[string]interface{}{
			"namaProyek": "",
		})
		assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	})

	t.Run("UpdateNotFound", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/000000000000000000000000",
			map[string]interface{}{"lokasi": "Somewhere"})
		if status != http.StatusNotFound {
			t.Fatalf("status = %d, want 404, body = %+v", status, env)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{"namaProyek": "THETA"})
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodDelete, "/api/v1/projects/"+id, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}

		status, env = apiRequest(t, baseURL, http.MethodGet, "/api/v1/projects/"+id, nil)
		if status != http.StatusNotFound {
			t.Fatalf("status after delete = %d, want 404, body = %+v", status, env)
		}
	})

	// A project holding items can't be deleted: the pre-v6 behavior left those
	// items with a dangling idProyek and no way to notice or fix it.
	t.Run("DeleteBlockedByReferencingItem", func(t *testing.T) {
		project := createProject(t, baseURL, map[string]interface{}{"namaProyek": "IOTA"})
		id := project["_id"].(string)
		item := createItem(t, baseURL, map[string]interface{}{"idProyek": id})

		status, env := apiRequest(t, baseURL, http.MethodDelete, "/api/v1/projects/"+id, nil)
		assertErrorEnvelope(t, status, env, http.StatusConflict, "PROJECT_IN_USE")
		details, ok := env["details"].(map[string]interface{})
		if !ok {
			t.Fatalf("details missing: %+v", env)
		}
		if details["itemCount"].(float64) != 1 {
			t.Errorf("details.itemCount = %v, want 1", details["itemCount"])
		}
		if details["namaProyek"] != "IOTA" {
			t.Errorf("details.namaProyek = %v, want IOTA", details["namaProyek"])
		}

		// Refused delete, no side effects: both project and item still there.
		if status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/projects/"+id, nil); status != http.StatusOK {
			t.Errorf("project should survive the refused delete: status = %d, body = %+v", status, env)
		}
		if status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/"+item["_id"].(string), nil); status != http.StatusOK {
			t.Errorf("item should be untouched by the refused delete: status = %d, body = %+v", status, env)
		}
	})

	// The way out: move the item to another project, then the delete works.
	t.Run("DeleteSucceedsAfterItemsMoveAway", func(t *testing.T) {
		from := createProject(t, baseURL, map[string]interface{}{"namaProyek": "LAMBDA-FROM"})
		to := createProject(t, baseURL, map[string]interface{}{"namaProyek": "LAMBDA-TO"})
		item := createItem(t, baseURL, map[string]interface{}{"idProyek": from["_id"].(string)})

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+item["_id"].(string),
			map[string]interface{}{"idProyek": to["_id"].(string)})
		if status != http.StatusOK {
			t.Fatalf("move item: status = %d, body = %+v", status, env)
		}

		status, env = apiRequest(t, baseURL, http.MethodDelete, "/api/v1/projects/"+from["_id"].(string), nil)
		if status != http.StatusOK {
			t.Fatalf("delete should succeed once the project is empty: status = %d, body = %+v", status, env)
		}
		assertSuccessCode(t, env, http.StatusOK)
		if nama := env["data"].(map[string]interface{})["namaProyek"]; nama != "LAMBDA-FROM" {
			t.Errorf("data.namaProyek = %v, want LAMBDA-FROM echoed back", nama)
		}
	})
}

// TestProjectsNamaProyekUniqueness exercises knowledge-base.md decision #12:
// namaProyek must be unique (case-sensitive, exact match) on both create and
// update, with update excluding the project's own document from the check.
func TestProjectsNamaProyekUniqueness(t *testing.T) {
	baseURL := setupBackendAPI(t)
	createProject(t, baseURL, map[string]interface{}{"namaProyek": "KAPPA"})

	t.Run("CreateDuplicateRejected", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/projects", map[string]interface{}{
			"namaProyek": "KAPPA",
		})
		assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		assertErrorFields(t, env, "namaProyek")
	})

	t.Run("CreateDifferentCaseAllowed", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/projects", map[string]interface{}{
			"namaProyek": "kappa",
		})
		if status != http.StatusCreated {
			t.Fatalf("case-sensitive uniqueness should allow differently-cased name: status = %d, body = %+v", status, env)
		}
	})

	t.Run("UpdateToDuplicateRejected", func(t *testing.T) {
		other := createProject(t, baseURL, map[string]interface{}{"namaProyek": "LAMBDA"})
		id := other["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/"+id, map[string]interface{}{
			"namaProyek": "KAPPA",
		})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422, body = %+v", status, env)
		}
	})

	t.Run("UpdateKeepingOwnNameAllowed", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{"namaProyek": "MU", "lokasi": "A"})
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/"+id, map[string]interface{}{
			"namaProyek": "MU", "lokasi": "B",
		})
		if status != http.StatusOK {
			t.Fatalf("re-submitting own unchanged name should not be treated as duplicate: status = %d, body = %+v", status, env)
		}
	})
}

// TestItemTypesLifecycle exercises api-contract.md §Jenis Barang: Create/
// List/Delete for the item-types master data collection, its
// case-insensitive uniqueness (unlike namaProyek), and that deleting one
// never touches items still using that jenis string (no FK).
func TestItemTypesLifecycle(t *testing.T) {
	baseURL := setupBackendAPI(t)

	t.Run("CreateAndList", func(t *testing.T) {
		createItemType(t, baseURL, "Drone")

		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/item-types", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		found := false
		for _, raw := range data {
			if raw.(map[string]interface{})["jenis"] == "Drone" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected Drone in list, got %+v", data)
		}
	})

	t.Run("CreateBlankRejected", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/item-types", map[string]interface{}{"jenis": "   "})
		assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		assertErrorFields(t, env, "jenis")
	})

	t.Run("CreateDuplicateRejected", func(t *testing.T) {
		createItemType(t, baseURL, "Printer")

		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/item-types", map[string]interface{}{"jenis": "Printer"})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422, body = %+v", status, env)
		}
	})

	t.Run("CreateDuplicateCaseInsensitiveRejected", func(t *testing.T) {
		createItemType(t, baseURL, "Scanner")

		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/item-types", map[string]interface{}{"jenis": "scanner"})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("case-insensitive uniqueness should reject differently-cased duplicate: status = %d, body = %+v", status, env)
		}
	})

	t.Run("DeleteRemovesFromList", func(t *testing.T) {
		created := createItemType(t, baseURL, "Router")
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodDelete, "/api/v1/item-types/"+id, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}

		_, listEnv := apiRequest(t, baseURL, http.MethodGet, "/api/v1/item-types", nil)
		for _, raw := range listEnv["data"].([]interface{}) {
			if raw.(map[string]interface{})["_id"] == id {
				t.Errorf("expected %s to be removed from list after delete", id)
			}
		}
	})

	t.Run("DeleteNotFound", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodDelete, "/api/v1/item-types/000000000000000000000000", nil)
		if status != http.StatusNotFound {
			t.Fatalf("status = %d, want 404, body = %+v", status, env)
		}
	})

	// Deleting a type that items still use is refused outright, and the item
	// documents are left completely untouched — no reassignment, no rewrite of
	// their jenis. The blocking count travels in details so the UI can name a
	// number without querying for it.
	t.Run("DeleteBlockedWhileItemsUseIt", func(t *testing.T) {
		created := createItemType(t, baseURL, "Tablet")
		id := created["_id"].(string)
		item := createItem(t, baseURL, map[string]interface{}{"jenis": "Tablet"})

		status, env := apiRequest(t, baseURL, http.MethodDelete, "/api/v1/item-types/"+id, nil)
		assertErrorEnvelope(t, status, env, http.StatusConflict, "ITEM_TYPE_IN_USE")
		details, ok := env["details"].(map[string]interface{})
		if !ok {
			t.Fatalf("details missing: %+v", env)
		}
		if details["itemCount"].(float64) != 1 {
			t.Errorf("details.itemCount = %v, want 1", details["itemCount"])
		}
		if details["jenis"] != "Tablet" {
			t.Errorf("details.jenis = %v, want Tablet", details["jenis"])
		}

		// The item keeps its jenis: a refused delete must not have side effects.
		_, itemEnv := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/"+item["_id"].(string), nil)
		if jenis := itemEnv["data"].(map[string]interface{})["jenis"]; jenis != "Tablet" {
			t.Errorf("item.jenis = %v, want it untouched at Tablet", jenis)
		}

		// And the type is still there to try again later.
		_, listEnv := apiRequest(t, baseURL, http.MethodGet, "/api/v1/item-types", nil)
		found := false
		for _, raw := range listEnv["data"].([]interface{}) {
			if raw.(map[string]interface{})["_id"] == id {
				found = true
			}
		}
		if !found {
			t.Errorf("expected Tablet to survive the refused delete, got %+v", listEnv["data"])
		}
	})

	// The escape hatch from the block above: move the last item off the type,
	// then the delete goes through. This is the flow the frontend has to guide
	// the user through, so it's worth pinning down end to end.
	t.Run("DeleteSucceedsAfterItemsMoveOff", func(t *testing.T) {
		createItemType(t, baseURL, "Plotter")
		_, listEnv := apiRequest(t, baseURL, http.MethodGet, "/api/v1/item-types", nil)
		var id string
		for _, raw := range listEnv["data"].([]interface{}) {
			row := raw.(map[string]interface{})
			if row["jenis"] == "Plotter" {
				id = row["_id"].(string)
			}
		}
		item := createItem(t, baseURL, map[string]interface{}{"jenis": "Plotter"})

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+item["_id"].(string),
			map[string]interface{}{"jenis": "Laptop"})
		if status != http.StatusOK {
			t.Fatalf("move item off the type: status = %d, body = %+v", status, env)
		}

		status, env = apiRequest(t, baseURL, http.MethodDelete, "/api/v1/item-types/"+id, nil)
		if status != http.StatusOK {
			t.Fatalf("delete should succeed once nothing uses the type: status = %d, body = %+v", status, env)
		}
		assertSuccessCode(t, env, http.StatusOK)
		if jenis := env["data"].(map[string]interface{})["jenis"]; jenis != "Plotter" {
			t.Errorf("data.jenis = %v, want Plotter echoed back for the confirmation toast", jenis)
		}
	})

	// "Laptop" and "laptop" are one type as far as uniqueness goes, so an item
	// written with either spelling has to block deleting it.
	t.Run("DeleteBlockedByDifferentlyCasedItems", func(t *testing.T) {
		created := createItemType(t, baseURL, "Kamera")
		id := created["_id"].(string)
		createItem(t, baseURL, map[string]interface{}{"jenis": "kamera"})

		status, env := apiRequest(t, baseURL, http.MethodDelete, "/api/v1/item-types/"+id, nil)
		assertErrorEnvelope(t, status, env, http.StatusConflict, "ITEM_TYPE_IN_USE")
	})

	// "Lainnya" lost its special status when reassign-on-delete went away: it
	// is now an ordinary entry, deletable like any other unused type.
	t.Run("LainnyaIsAnOrdinaryType", func(t *testing.T) {
		created := createItemType(t, baseURL, "Lainnya")

		status, env := apiRequest(t, baseURL, http.MethodDelete, "/api/v1/item-types/"+created["_id"].(string), nil)
		if status != http.StatusOK {
			t.Fatalf("Lainnya should be deletable like any unused type: status = %d, body = %+v", status, env)
		}
	})
}

// TestItemTypeAutoRegistration covers the other half of the "every jenis in
// use is in the master list" invariant: writing an item with a brand-new
// jenis registers that type, without the client calling POST /item-types.
func TestItemTypeAutoRegistration(t *testing.T) {
	baseURL := setupBackendAPI(t)

	listJenis := func() []string {
		t.Helper()
		_, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/filter-options/jenis", nil)
		values := make([]string, 0)
		for _, raw := range env["data"].([]interface{}) {
			values = append(values, raw.(string))
		}
		return values
	}

	contains := func(values []string, want string) bool {
		for _, v := range values {
			if v == want {
				return true
			}
		}
		return false
	}

	t.Run("CreateItemRegistersNewJenis", func(t *testing.T) {
		createItem(t, baseURL, map[string]interface{}{"jenis": "Starlink"})
		if values := listJenis(); !contains(values, "Starlink") {
			t.Errorf("jenis options = %v, want Starlink registered by the item create", values)
		}
	})

	t.Run("UpdateItemRegistersNewJenis", func(t *testing.T) {
		item := createItem(t, baseURL, nil)
		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+item["_id"].(string),
			map[string]interface{}{"jenis": "Genset"})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		if values := listJenis(); !contains(values, "Genset") {
			t.Errorf("jenis options = %v, want Genset registered by the item update", values)
		}
	})

	t.Run("ExistingJenisNotDuplicated", func(t *testing.T) {
		createItem(t, baseURL, map[string]interface{}{"jenis": "Kamera"})
		createItem(t, baseURL, map[string]interface{}{"jenis": "kamera"})

		count := 0
		for _, v := range listJenis() {
			if strings.EqualFold(v, "Kamera") {
				count++
			}
		}
		if count != 1 {
			t.Errorf("found %d entries for Kamera, want 1 (matching is case-insensitive)", count)
		}
	})
}

// TestMalformedBody exercises the 400 MALFORMED_BODY path shared across
// every mutating endpoint: an empty body, invalid JSON, or JSON that isn't
// an object.
func TestMalformedBody(t *testing.T) {
	baseURL := setupBackendAPI(t)

	postRaw := func(t *testing.T, path, rawBody string) (int, map[string]interface{}) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewBufferString(rawBody))
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

	assertMalformed := func(t *testing.T, status int, env map[string]interface{}) {
		t.Helper()
		assertErrorEnvelope(t, status, env, http.StatusBadRequest, "MALFORMED_BODY")
	}

	t.Run("EmptyBodyOnCreateItem", func(t *testing.T) {
		status, env := postRaw(t, "/api/v1/items", "")
		assertMalformed(t, status, env)
	})

	t.Run("InvalidJSONOnCreateItem", func(t *testing.T) {
		status, env := postRaw(t, "/api/v1/items", "{not valid json")
		assertMalformed(t, status, env)
	})

	t.Run("JSONArrayOnCreateItem", func(t *testing.T) {
		status, env := postRaw(t, "/api/v1/items", "[1,2,3]")
		assertMalformed(t, status, env)
	})

	t.Run("EmptyBodyOnCreateProject", func(t *testing.T) {
		status, env := postRaw(t, "/api/v1/projects", "")
		assertMalformed(t, status, env)
	})

	t.Run("EmptyObjectIsNotMalformed", func(t *testing.T) {
		// A valid-but-empty JSON object is not MALFORMED_BODY — it should
		// fall through to normal required-field validation (422), not 400.
		status, env := postRaw(t, "/api/v1/items", "{}")
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want 422 (empty object is valid JSON, just missing required fields), body = %+v", status, env)
		}
	})
}

// TestListLimitClamp checks that GET /items?limit=... is clamped to a
// maximum of 100 rather than honored as-is or rejected.
func TestListLimitClamp(t *testing.T) {
	baseURL := setupBackendAPI(t)

	status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?limit=500", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, body = %+v", status, env)
	}
	meta := env["meta"].(map[string]interface{})
	if meta["limit"].(float64) != 100 {
		t.Errorf("meta.limit = %v, want clamped to 100", meta["limit"])
	}
}

// TestItemNamaProyekField checks that every Item response path (create,
// get, list, update) denormalizes idProyek into a namaProyek the client
// can display without a separate /projects lookup, and that it degrades
// gracefully (empty, not an error) for an orphaned idProyek.
func TestItemNamaProyekField(t *testing.T) {
	baseURL := setupBackendAPI(t)

	project := createProject(t, baseURL, map[string]interface{}{"namaProyek": "PROJNAME-TEST"})
	projectID := project["_id"].(string)

	t.Run("Create", func(t *testing.T) {
		created := createItem(t, baseURL, map[string]interface{}{"idProyek": projectID})
		if created["namaProyek"] != "PROJNAME-TEST" {
			t.Errorf("namaProyek = %v, want PROJNAME-TEST", created["namaProyek"])
		}
	})

	t.Run("Get", func(t *testing.T) {
		created := createItem(t, baseURL, map[string]interface{}{"idProyek": projectID})
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/"+created["_id"].(string), nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["namaProyek"] != "PROJNAME-TEST" {
			t.Errorf("namaProyek = %v, want PROJNAME-TEST", data["namaProyek"])
		}
	})

	t.Run("List", func(t *testing.T) {
		createItem(t, baseURL, map[string]interface{}{"idProyek": projectID})
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?idProyek="+projectID, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected at least one item")
		}
		first := data[0].(map[string]interface{})
		if first["namaProyek"] != "PROJNAME-TEST" {
			t.Errorf("namaProyek = %v, want PROJNAME-TEST", first["namaProyek"])
		}
	})

	t.Run("Update", func(t *testing.T) {
		created := createItem(t, baseURL, map[string]interface{}{"idProyek": projectID})
		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+created["_id"].(string),
			map[string]interface{}{"status": "Broken"})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["namaProyek"] != "PROJNAME-TEST" {
			t.Errorf("namaProyek = %v, want PROJNAME-TEST", data["namaProyek"])
		}
	})

	t.Run("SendingNamaProyekOnCreateIsIgnoredNotFoldedIntoCustomAttributes", func(t *testing.T) {
		created := createItem(t, baseURL, map[string]interface{}{
			"idProyek":   projectID,
			"namaProyek": "attacker-supplied",
		})
		if created["namaProyek"] != "PROJNAME-TEST" {
			t.Errorf("namaProyek = %v, want server-resolved PROJNAME-TEST (client value must be ignored)", created["namaProyek"])
		}
		if custom, ok := created["customAttributes"].(map[string]interface{}); ok {
			if _, ok := custom["namaProyek"]; ok {
				t.Errorf("namaProyek leaked into customAttributes: %+v", custom)
			}
		}
	})

}

// TestLegacyOrphanedIdProyek covers reading data that predates the v6 delete
// block: an item whose project is simply gone. The API can no longer produce
// this state (DELETE /projects is refused while items reference it), so the
// orphan is manufactured by deleting the project document directly — the same
// shape a database carried over from v5 can still hold.
func TestLegacyOrphanedIdProyek(t *testing.T) {
	baseURL, database := setupBackendAPIWithDB(t)

	project := createProject(t, baseURL, map[string]interface{}{"namaProyek": "ORPHAN-SRC"})
	item := createItem(t, baseURL, map[string]interface{}{"idProyek": project["_id"]})

	if _, err := database.Collection("projects").DeleteOne(context.Background(),
		bson.M{"_id": project["_id"].(string)}); err != nil {
		t.Fatalf("failed to simulate legacy orphan: %v", err)
	}

	t.Run("NamaProyekIsEmptyNotAnError", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/"+item["_id"].(string), nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if v, ok := data["namaProyek"]; ok && v != "" {
			t.Errorf("namaProyek = %v, want empty/absent for orphaned idProyek", v)
		}
	})

	// The dashboard's way of finding out it has legacy junk to clean up.
	t.Run("CountedInSummaryNeedsAttention", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/analytics/summary", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		needsAttention := env["data"].(map[string]interface{})["needsAttention"].(map[string]interface{})
		if needsAttention["orphanItems"].(float64) != 1 {
			t.Errorf("orphanItems = %v, want 1", needsAttention["orphanItems"])
		}
	})
}

// TestProjectKoordinat covers the map coordinate field added in v6: it's
// optional, all-or-nothing, range-checked, replaced wholesale on PATCH, and
// removable with an explicit null.
func TestProjectKoordinat(t *testing.T) {
	baseURL := setupBackendAPI(t)

	assertPoint := func(t *testing.T, data map[string]interface{}, wantLat, wantLng float64) {
		t.Helper()
		koordinat, ok := data["koordinat"].(map[string]interface{})
		if !ok {
			t.Fatalf("koordinat missing or wrong type: %+v", data)
		}
		if koordinat["lat"].(float64) != wantLat || koordinat["lng"].(float64) != wantLng {
			t.Errorf("koordinat = %+v, want lat=%v lng=%v", koordinat, wantLat, wantLng)
		}
	}

	t.Run("CreateWithKoordinat", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{
			"namaProyek": "KOORD-1",
			"koordinat":  map[string]interface{}{"lat": -6.2088, "lng": 106.8456},
		})
		assertPoint(t, created, -6.2088, 106.8456)
	})

	// A project without a point is valid — it just can't be pinned yet. The
	// key is absent rather than null so clients can test for presence.
	t.Run("CreateWithoutKoordinatOmitsTheField", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{"namaProyek": "KOORD-NONE"})
		if _, ok := created["koordinat"]; ok {
			t.Errorf("koordinat should be absent, got %+v", created["koordinat"])
		}
	})

	// Forms hand back input values as strings; requiring the client to convert
	// them first would be gratuitous.
	t.Run("NumericStringsAccepted", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{
			"namaProyek": "KOORD-STR",
			"koordinat":  map[string]interface{}{"lat": "-7.2575", "lng": "112.7521"},
		})
		assertPoint(t, created, -7.2575, 112.7521)
	})

	t.Run("HalfACoordinateRejected", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/projects", map[string]interface{}{
			"namaProyek": "KOORD-HALF",
			"koordinat":  map[string]interface{}{"lat": -6.2},
		})
		assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		assertErrorFields(t, env, "koordinat.lng")
	})

	// An empty object is a bug signature (a form serializing undefined values),
	// not a request to clear the point — so it fails instead of silently
	// wiping a correct coordinate.
	t.Run("EmptyObjectRejected", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/projects", map[string]interface{}{
			"namaProyek": "KOORD-EMPTY",
			"koordinat":  map[string]interface{}{},
		})
		assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		assertErrorFields(t, env, "koordinat.lat", "koordinat.lng")
	})

	t.Run("OutOfRangeRejected", func(t *testing.T) {
		cases := []struct {
			name  string
			lat   interface{}
			lng   interface{}
			field string
		}{
			{"LatTooHigh", 91.0, 100.0, "koordinat.lat"},
			{"LatTooLow", -91.0, 100.0, "koordinat.lat"},
			{"LngTooHigh", 0.0, 181.0, "koordinat.lng"},
			{"LngTooLow", 0.0, -181.0, "koordinat.lng"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/projects", map[string]interface{}{
					"namaProyek": "KOORD-RANGE-" + tc.name,
					"koordinat":  map[string]interface{}{"lat": tc.lat, "lng": tc.lng},
				})
				assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
				assertErrorFields(t, env, tc.field)
			})
		}
	})

	t.Run("NonNumericRejected", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/projects", map[string]interface{}{
			"namaProyek": "KOORD-NAN",
			"koordinat":  map[string]interface{}{"lat": "abc", "lng": ""},
		})
		assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		assertErrorFields(t, env, "koordinat.lat", "koordinat.lng")
	})

	t.Run("NonObjectRejected", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPost, "/api/v1/projects", map[string]interface{}{
			"namaProyek": "KOORD-STRING",
			"koordinat":  "-6.2,106.8",
		})
		assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		assertErrorFields(t, env, "koordinat")
	})

	// Top-level latitude/longitude are not part of the contract; like any other
	// unknown project key they're ignored (Project has no customAttributes).
	t.Run("FlatLatitudeLongitudeIgnored", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{
			"namaProyek": "KOORD-FLAT",
			"latitude":   -6.2,
			"longitude":  106.8,
		})
		if _, ok := created["koordinat"]; ok {
			t.Errorf("flat latitude/longitude should not become koordinat: %+v", created)
		}
		if _, ok := created["latitude"]; ok {
			t.Errorf("unknown project keys should be dropped, got %+v", created)
		}
	})

	t.Run("PatchSetsAndMovesThePoint", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{"namaProyek": "KOORD-PATCH"})
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/"+id, map[string]interface{}{
			"koordinat": map[string]interface{}{"lat": -6.9175, "lng": 107.6191},
		})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		assertPoint(t, env["data"].(map[string]interface{}), -6.9175, 107.6191)

		status, env = apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/"+id, map[string]interface{}{
			"koordinat": map[string]interface{}{"lat": -5.1477, "lng": 119.4327},
		})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		assertPoint(t, env["data"].(map[string]interface{}), -5.1477, 119.4327)
	})

	// koordinat is untouched by a PATCH that doesn't mention it — same rule as
	// any other field.
	t.Run("PatchWithoutKoordinatKeepsIt", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{
			"namaProyek": "KOORD-KEEP",
			"koordinat":  map[string]interface{}{"lat": 3.5952, "lng": 98.6722},
		})
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/"+id,
			map[string]interface{}{"lokasi": "Medan Warehouse 2"})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		assertPoint(t, env["data"].(map[string]interface{}), 3.5952, 98.6722)
	})

	// There's no per-component patch: a partial object is an error, not a
	// nudge to one axis.
	t.Run("PatchPartialKoordinatRejected", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{
			"namaProyek": "KOORD-PARTIAL",
			"koordinat":  map[string]interface{}{"lat": -6.2, "lng": 106.8},
		})
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/"+id, map[string]interface{}{
			"koordinat": map[string]interface{}{"lat": -6.3},
		})
		assertErrorEnvelope(t, status, env, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		assertErrorFields(t, env, "koordinat.lng")

		// and the stored point is unchanged
		_, getEnv := apiRequest(t, baseURL, http.MethodGet, "/api/v1/projects/"+id, nil)
		assertPoint(t, getEnv["data"].(map[string]interface{}), -6.2, 106.8)
	})

	t.Run("PatchNullRemovesThePoint", func(t *testing.T) {
		created := createProject(t, baseURL, map[string]interface{}{
			"namaProyek": "KOORD-CLEAR",
			"koordinat":  map[string]interface{}{"lat": -6.2, "lng": 106.8},
		})
		id := created["_id"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/projects/"+id,
			map[string]interface{}{"koordinat": nil})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		if _, ok := env["data"].(map[string]interface{})["koordinat"]; ok {
			t.Errorf("koordinat should be gone after an explicit null, got %+v", env["data"])
		}

		_, getEnv := apiRequest(t, baseURL, http.MethodGet, "/api/v1/projects/"+id, nil)
		if _, ok := getEnv["data"].(map[string]interface{})["koordinat"]; ok {
			t.Errorf("koordinat should stay gone after re-read, got %+v", getEnv["data"])
		}
	})

	// The dropdown payload doubles as map data, so it has to carry the point.
	t.Run("FilterOptionsProjectCarriesKoordinat", func(t *testing.T) {
		createProject(t, baseURL, map[string]interface{}{
			"namaProyek": "KOORD-FILTEROPT",
			"koordinat":  map[string]interface{}{"lat": -1.2379, "lng": 116.8529},
		})

		_, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/filter-options", nil)
		for _, raw := range env["data"].(map[string]interface{})["project"].([]interface{}) {
			row := raw.(map[string]interface{})
			if row["namaProyek"] == "KOORD-FILTEROPT" {
				assertPoint(t, row, -1.2379, 116.8529)
				return
			}
		}
		t.Errorf("KOORD-FILTEROPT missing from filter-options project list")
	})
}

// TestAnalyticsSummary checks the dashboard numbers against a known dataset.
func TestAnalyticsSummary(t *testing.T) {
	baseURL := setupBackendAPI(t)

	mapped := createProject(t, baseURL, map[string]interface{}{
		"namaProyek": "AN-MAPPED",
		"lokasi":     "Jakarta HQ",
		"koordinat":  map[string]interface{}{"lat": -6.2088, "lng": 106.8456},
	})
	unmappedProject := createProject(t, baseURL, map[string]interface{}{"namaProyek": "AN-UNMAPPED"})
	createProject(t, baseURL, map[string]interface{}{"namaProyek": "AN-EMPTY"})

	createItem(t, baseURL, map[string]interface{}{
		"idProyek": mapped["_id"], "jenis": "Laptop", "status": "Healthy",
		"licenseWindows": "Pro", "licenseOffice": "365",
	})
	createItem(t, baseURL, map[string]interface{}{
		"idProyek": mapped["_id"], "jenis": "Laptop", "status": "Broken",
		"licenseWindows": "Pro", "licenseOffice": "",
	})
	createItem(t, baseURL, map[string]interface{}{
		"idProyek": unmappedProject["_id"], "jenis": "Monitor", "status": "Healthy",
	})

	status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/analytics/summary", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, body = %+v", status, env)
	}
	assertSuccessCode(t, env, http.StatusOK)
	data := env["data"].(map[string]interface{})

	t.Run("Totals", func(t *testing.T) {
		totals := data["totals"].(map[string]interface{})
		for field, want := range map[string]float64{
			"items": 3, "projects": 3, "projectsMapped": 1,
		} {
			if totals[field].(float64) != want {
				t.Errorf("totals.%s = %v, want %v", field, totals[field], want)
			}
		}
		// itemTypes counts the master list, which auto-registration filled in
		// from the two jenis actually used.
		if totals["itemTypes"].(float64) != 2 {
			t.Errorf("totals.itemTypes = %v, want 2", totals["itemTypes"])
		}
	})

	t.Run("ByStatusAndJenisSortedByCount", func(t *testing.T) {
		byStatus := data["byStatus"].([]interface{})
		first := byStatus[0].(map[string]interface{})
		if first["status"] != "Healthy" || first["count"].(float64) != 2 {
			t.Errorf("byStatus[0] = %+v, want Healthy x2 first", first)
		}
		byJenis := data["byJenis"].([]interface{})
		firstJenis := byJenis[0].(map[string]interface{})
		if firstJenis["jenis"] != "Laptop" || firstJenis["count"].(float64) != 2 {
			t.Errorf("byJenis[0] = %+v, want Laptop x2 first", firstJenis)
		}
	})

	// byProyek is pre-joined with project names, and empty projects are listed
	// too — otherwise "which sites are empty" needs a second request.
	t.Run("ByProyekIncludesNamesAndEmptyProjects", func(t *testing.T) {
		byProyek := data["byProyek"].([]interface{})
		if len(byProyek) != 3 {
			t.Fatalf("byProyek = %d entries, want 3 (one per project)", len(byProyek))
		}
		counts := map[string]float64{}
		for _, raw := range byProyek {
			row := raw.(map[string]interface{})
			counts[row["namaProyek"].(string)] = row["count"].(float64)
			if row["idProyek"] == "" {
				t.Errorf("byProyek entry missing idProyek: %+v", row)
			}
		}
		for nama, want := range map[string]float64{"AN-MAPPED": 2, "AN-UNMAPPED": 1, "AN-EMPTY": 0} {
			if counts[nama] != want {
				t.Errorf("byProyek[%s] = %v, want %v", nama, counts[nama], want)
			}
		}
	})

	// A blank license means "not applicable", so it isn't a chart category.
	t.Run("LicensesSkipBlanks", func(t *testing.T) {
		licenses := data["licenses"].(map[string]interface{})
		windows := licenses["windows"].([]interface{})
		if len(windows) != 1 {
			t.Fatalf("licenses.windows = %+v, want a single Pro bucket", windows)
		}
		if bucket := windows[0].(map[string]interface{}); bucket["value"] != "Pro" || bucket["count"].(float64) != 2 {
			t.Errorf("licenses.windows[0] = %+v, want Pro x2", bucket)
		}
		office := licenses["office"].([]interface{})
		if len(office) != 1 {
			t.Errorf("licenses.office = %+v, want only the non-blank 365 bucket", office)
		}
	})

	t.Run("NeedsAttention", func(t *testing.T) {
		needsAttention := data["needsAttention"].(map[string]interface{})
		for field, want := range map[string]float64{
			"staleDays": 90, "staleItems": 0, "orphanItems": 0,
			"projectsWithoutKoordinat": 2, "projectsWithoutItems": 1,
		} {
			if needsAttention[field].(float64) != want {
				t.Errorf("needsAttention.%s = %v, want %v", field, needsAttention[field], want)
			}
		}
	})

	// staleDays is a display knob: garbage input falls back to the default
	// rather than failing the request and blanking the dashboard.
	t.Run("StaleDaysParamHonoredAndClamped", func(t *testing.T) {
		for _, tc := range []struct {
			query string
			want  float64
		}{
			{"?staleDays=30", 30},
			{"?staleDays=99999", 3650},
			{"?staleDays=abc", 90},
			{"?staleDays=0", 1},
		} {
			_, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/analytics/summary"+tc.query, nil)
			got := env["data"].(map[string]interface{})["needsAttention"].(map[string]interface{})["staleDays"].(float64)
			if got != tc.want {
				t.Errorf("%s -> staleDays = %v, want %v", tc.query, got, tc.want)
			}
		}
	})

	t.Run("RecentListsCarryFullItems", func(t *testing.T) {
		for _, key := range []string{"recentlyAdded", "recentlyUpdated"} {
			list := data[key].([]interface{})
			if len(list) != 3 {
				t.Fatalf("%s = %d items, want 3", key, len(list))
			}
			first := list[0].(map[string]interface{})
			for _, field := range []string{"_id", "jenis", "serialNumber", "status", "idProyek", "namaProyek"} {
				if _, ok := first[field]; !ok {
					t.Errorf("%s[0] missing %q: %+v", key, field, first)
				}
			}
		}
	})
}

// TestAnalyticsMap covers the map layer payload: only mapped projects become
// pins, unmapped ones are reported separately rather than dropped, and bounds
// is null when there's nothing to fit.
func TestAnalyticsMap(t *testing.T) {
	baseURL := setupBackendAPI(t)

	t.Run("EmptyDatabase", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/analytics/map", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if len(data["projects"].([]interface{})) != 0 || len(data["unmapped"].([]interface{})) != 0 {
			t.Errorf("expected empty projects/unmapped, got %+v", data)
		}
		if data["bounds"] != nil {
			t.Errorf("bounds = %+v, want null when nothing is mapped", data["bounds"])
		}
	})

	north := createProject(t, baseURL, map[string]interface{}{
		"namaProyek": "MAP-NORTH", "lokasi": "Medan",
		"koordinat": map[string]interface{}{"lat": 3.5952, "lng": 98.6722},
	})
	south := createProject(t, baseURL, map[string]interface{}{
		"namaProyek": "MAP-SOUTH", "lokasi": "Surabaya",
		"koordinat": map[string]interface{}{"lat": -7.2575, "lng": 112.7521},
	})
	createProject(t, baseURL, map[string]interface{}{"namaProyek": "MAP-NOPOINT", "lokasi": "Semarang"})

	createItem(t, baseURL, map[string]interface{}{"idProyek": north["_id"], "status": "Healthy"})
	createItem(t, baseURL, map[string]interface{}{"idProyek": north["_id"], "status": "Healthy"})
	createItem(t, baseURL, map[string]interface{}{"idProyek": north["_id"], "status": "Broken"})
	createItem(t, baseURL, map[string]interface{}{"idProyek": south["_id"], "status": "Broken"})

	status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/analytics/map", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, body = %+v", status, env)
	}
	data := env["data"].(map[string]interface{})

	t.Run("MappedProjectsCarryCountsAndStatusMix", func(t *testing.T) {
		projects := data["projects"].([]interface{})
		if len(projects) != 2 {
			t.Fatalf("projects = %d, want 2 mapped", len(projects))
		}
		// sorted by namaProyek: MAP-NORTH before MAP-SOUTH
		first := projects[0].(map[string]interface{})
		if first["namaProyek"] != "MAP-NORTH" {
			t.Fatalf("projects[0] = %v, want MAP-NORTH (sorted by name)", first["namaProyek"])
		}
		if first["totalItems"].(float64) != 3 {
			t.Errorf("MAP-NORTH totalItems = %v, want 3", first["totalItems"])
		}
		if first["dominantStatus"] != "Healthy" {
			t.Errorf("MAP-NORTH dominantStatus = %v, want Healthy", first["dominantStatus"])
		}
		byStatus := first["byStatus"].([]interface{})
		if len(byStatus) != 2 {
			t.Fatalf("MAP-NORTH byStatus = %+v, want 2 buckets", byStatus)
		}
		if top := byStatus[0].(map[string]interface{}); top["status"] != "Healthy" || top["count"].(float64) != 2 {
			t.Errorf("byStatus[0] = %+v, want Healthy x2", top)
		}
		if _, ok := first["koordinat"].(map[string]interface{}); !ok {
			t.Errorf("mapped project missing koordinat: %+v", first)
		}
	})

	t.Run("UnmappedProjectsReportedNotDropped", func(t *testing.T) {
		unmapped := data["unmapped"].([]interface{})
		if len(unmapped) != 1 {
			t.Fatalf("unmapped = %+v, want 1 entry", unmapped)
		}
		row := unmapped[0].(map[string]interface{})
		if row["namaProyek"] != "MAP-NOPOINT" {
			t.Errorf("unmapped[0].namaProyek = %v, want MAP-NOPOINT", row["namaProyek"])
		}
		if _, ok := row["koordinat"]; ok {
			t.Errorf("unmapped entry should have no koordinat: %+v", row)
		}
	})

	t.Run("Bounds", func(t *testing.T) {
		bounds := data["bounds"].(map[string]interface{})
		for field, want := range map[string]float64{
			"north": 3.5952, "south": -7.2575, "east": 112.7521, "west": 98.6722,
		} {
			if bounds[field].(float64) != want {
				t.Errorf("bounds.%s = %v, want %v", field, bounds[field], want)
			}
		}
	})

	t.Run("Totals", func(t *testing.T) {
		totals := data["totals"].(map[string]interface{})
		for field, want := range map[string]float64{
			"projects": 3, "mapped": 2, "unmapped": 1, "items": 4, "itemsMapped": 4,
		} {
			if totals[field].(float64) != want {
				t.Errorf("totals.%s = %v, want %v", field, totals[field], want)
			}
		}
	})
}

// TestAnalyticsTimeline checks the month buckets are contiguous, zero-filled,
// and cumulative across the whole collection.
func TestAnalyticsTimeline(t *testing.T) {
	baseURL := setupBackendAPI(t)

	createItem(t, baseURL, nil)
	createItem(t, baseURL, nil)

	t.Run("DefaultWindow", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/analytics/timeline", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		assertSuccessCode(t, env, http.StatusOK)
		data := env["data"].(map[string]interface{})
		if data["months"].(float64) != 12 {
			t.Errorf("months = %v, want the default 12", data["months"])
		}
		buckets := data["buckets"].([]interface{})
		if len(buckets) != 12 {
			t.Fatalf("buckets = %d, want exactly 12 (zero-filled)", len(buckets))
		}
		if data["from"] != buckets[0].(map[string]interface{})["period"] {
			t.Errorf("from = %v, want the first bucket's period", data["from"])
		}
		if data["to"] != buckets[11].(map[string]interface{})["period"] {
			t.Errorf("to = %v, want the last bucket's period", data["to"])
		}

		// Both items were just created, so they land in the final (current)
		// month, and the running total ends at the collection size.
		last := buckets[11].(map[string]interface{})
		if last["created"].(float64) != 2 {
			t.Errorf("last bucket created = %v, want 2", last["created"])
		}
		if last["cumulative"].(float64) != 2 {
			t.Errorf("last bucket cumulative = %v, want 2", last["cumulative"])
		}

		// Contiguity: every bucket present, monotonic cumulative.
		prev := 0.0
		for i, raw := range buckets {
			bucket := raw.(map[string]interface{})
			if _, ok := bucket["period"].(string); !ok {
				t.Fatalf("bucket %d has no period: %+v", i, bucket)
			}
			cumulative := bucket["cumulative"].(float64)
			if cumulative < prev {
				t.Errorf("cumulative went down at bucket %d: %v after %v", i, cumulative, prev)
			}
			prev = cumulative
		}
	})

	t.Run("MonthsParamHonoredAndClamped", func(t *testing.T) {
		for _, tc := range []struct {
			query string
			want  int
		}{
			{"?months=3", 3},
			{"?months=1", 1},
			{"?months=999", 60},
			{"?months=abc", 12},
			{"?months=0", 1},
		} {
			_, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/analytics/timeline"+tc.query, nil)
			data := env["data"].(map[string]interface{})
			if got := int(data["months"].(float64)); got != tc.want {
				t.Errorf("%s -> months = %d, want %d", tc.query, got, tc.want)
			}
			if got := len(data["buckets"].([]interface{})); got != tc.want {
				t.Errorf("%s -> %d buckets, want %d", tc.query, got, tc.want)
			}
		}
	})
}

// TestSearchTreatsQAsLiteralText pins down that `q` is a substring search, not
// a regex the caller can inject: an unescaped needle used to reach Mongo
// directly, so "(" failed the whole query with a 500 and ".*" matched every
// item instead of those two characters.
func TestSearchTreatsQAsLiteralText(t *testing.T) {
	baseURL := setupBackendAPI(t)

	project := createProject(t, baseURL, nil)
	createItem(t, baseURL, map[string]interface{}{
		"idProyek": project["_id"], "nama": "RTI-(SPECIAL)-001", "serialNumber": "SN-REGEX-001",
	})
	createItem(t, baseURL, map[string]interface{}{
		"idProyek": project["_id"], "nama": "RTI-PLAIN-002", "serialNumber": "SN-REGEX-002",
	})

	search := func(t *testing.T, q string) (int, float64) {
		t.Helper()
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items?q="+url.QueryEscape(q), nil)
		if status != http.StatusOK {
			t.Fatalf("q=%q: status = %d, want 200, body = %+v", q, status, env)
		}
		return status, env["meta"].(map[string]interface{})["total"].(float64)
	}

	// Metacharacters are matched literally, and an unbalanced one is an
	// ordinary no-match rather than a server error.
	for _, tc := range []struct {
		q    string
		want float64
	}{
		{"(SPECIAL)", 1},
		{"(", 1},
		{")", 1},
		{".*", 0},
		{"[a-z]", 0},
		{"RTI-", 2},
	} {
		if _, total := search(t, tc.q); total != tc.want {
			t.Errorf("q=%q matched %v item(s), want %v", tc.q, total, tc.want)
		}
	}
}

// TestServerAssignedItemFieldsIgnored covers the read-only half of the item
// shape: _id and the timestamps are the server's to set. createdAt in
// particular used to be writable through PATCH, which let a client backdate an
// item and skew the analytics timeline built from that field.
func TestServerAssignedItemFieldsIgnored(t *testing.T) {
	baseURL := setupBackendAPI(t)
	const fakeID = "deadbeefdeadbeefdeadbeef"
	const fakeTime = "1999-01-01T00:00:00.000Z"

	t.Run("CreateIgnoresThem", func(t *testing.T) {
		created := createItem(t, baseURL, map[string]interface{}{
			"_id": fakeID, "createdAt": fakeTime, "updatedAt": fakeTime,
		})
		if created["_id"] == fakeID {
			t.Errorf("_id = %v, want a server-generated id", created["_id"])
		}
		if created["createdAt"] == fakeTime || created["updatedAt"] == fakeTime {
			t.Errorf("timestamps = %v/%v, want server-assigned", created["createdAt"], created["updatedAt"])
		}
	})

	t.Run("PatchIgnoresThem", func(t *testing.T) {
		item := createItem(t, baseURL, nil)
		id := item["_id"].(string)
		originalCreatedAt := item["createdAt"].(string)

		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+id, map[string]interface{}{
			"_id": fakeID, "createdAt": fakeTime, "updatedAt": fakeTime, "status": "Broken",
		})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["_id"] != id {
			t.Errorf("_id = %v, want unchanged %v", data["_id"], id)
		}
		if data["createdAt"] != originalCreatedAt {
			t.Errorf("createdAt = %v, want unchanged %v", data["createdAt"], originalCreatedAt)
		}
		if data["updatedAt"] == fakeTime {
			t.Errorf("updatedAt = %v, want a fresh server stamp", data["updatedAt"])
		}
		if data["status"] != "Broken" {
			t.Errorf("status = %v, want the legitimate part of the patch to still apply", data["status"])
		}
		// and they must not have been quietly folded into customAttributes
		if custom, ok := data["customAttributes"].(map[string]interface{}); ok {
			for _, key := range []string{"_id", "createdAt", "updatedAt"} {
				if _, leaked := custom[key]; leaked {
					t.Errorf("%s leaked into customAttributes: %+v", key, custom)
				}
			}
		}
	})
}
