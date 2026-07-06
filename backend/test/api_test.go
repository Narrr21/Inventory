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

// setupBackendAPI wires up the real backend (Mongo connection, repository,
// handlers, chi router) exactly as cmd/server/main.go does, and serves it
// over a real HTTP listener so tests exercise the API as an outside client
// would. It connects to MONGODB_URI/MONGODB_TEST_DB (defaulting to a local
// instance and a dedicated test database), cleans the items collection
// before and after the test, and skips if no MongoDB instance is reachable.
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
	if _, err := itemsColl.DeleteMany(context.Background(), map[string]interface{}{}); err != nil {
		t.Fatalf("failed to clean items collection: %v", err)
	}

	repo := repository.NewItemRepository(client.Database)
	router := handlers.NewRouter(handlers.NewItemHandler(repo))
	server := httptest.NewServer(router)

	t.Cleanup(func() {
		server.Close()
		_, _ = itemsColl.DeleteMany(context.Background(), map[string]interface{}{})
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

func createItem(t *testing.T, baseURL string, overrides map[string]interface{}) map[string]interface{} {
	t.Helper()
	body := map[string]interface{}{
		"jenisProduct": "Laptop",
		"serialNumber": "SN-TEST-001",
		"name":         "RTI-TEST-001",
		"proyek":       "ALPHA",
		"status":       "Active",
		"lokasi":       "Jakarta HQ",
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
		created := createItem(t, baseURL, map[string]interface{}{"name": "RTI-LIFECYCLE-001"})
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
		if _, ok := first["passwordPin"]; ok {
			t.Error("expected sensitive fields to be omitted from list response")
		}
	})

	t.Run("Update", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodPatch, "/api/v1/items/"+itemID, map[string]interface{}{
			"status": "Maintenance",
			"lokasi": "Gudang",
		})
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		if data["status"] != "Maintenance" || data["lokasi"] != "Gudang" {
			t.Errorf("data = %+v, want status=Maintenance lokasi=Gudang", data)
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

// TestFilterOptionsAndStats exercises the aggregate endpoints against seeded data.
func TestFilterOptionsAndStats(t *testing.T) {
	baseURL := setupBackendAPI(t)

	createItem(t, baseURL, map[string]interface{}{
		"name": "RTI-AGG-001", "proyek": "ALPHA", "jenisProduct": "Laptop", "lokasi": "Jakarta HQ", "status": "Active",
	})
	createItem(t, baseURL, map[string]interface{}{
		"name": "RTI-AGG-002", "proyek": "BETA", "jenisProduct": "Monitor", "lokasi": "Surabaya Branch", "status": "Idle",
	})

	t.Run("FilterOptions", func(t *testing.T) {
		status, env := apiRequest(t, baseURL, http.MethodGet, "/api/v1/items/filter-options", nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %+v", status, env)
		}
		data := env["data"].(map[string]interface{})
		proyek := data["proyek"].([]interface{})
		if len(proyek) != 2 {
			t.Errorf("proyek = %v, want 2 distinct values", proyek)
		}
	})

	t.Run("FilterOptionsPerField", func(t *testing.T) {
		cases := []struct {
			path string
			want []string
		}{
			{"/api/v1/items/filter-options/lokasi", []string{"Jakarta HQ", "Surabaya Branch"}},
			{"/api/v1/items/filter-options/proyek", []string{"ALPHA", "BETA"}},
			{"/api/v1/items/filter-options/jenisProduct", []string{"Laptop", "Monitor"}},
		}
		for _, tc := range cases {
			status, env := apiRequest(t, baseURL, http.MethodGet, tc.path, nil)
			if status != http.StatusOK {
				t.Fatalf("%s: status = %d, body = %+v", tc.path, status, env)
			}
			data := env["data"].([]interface{})
			if len(data) != len(tc.want) {
				t.Errorf("%s: data = %v, want %v", tc.path, data, tc.want)
			}
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
		recent := data["recentlyAdded"].([]interface{})
		if len(recent) != 2 {
			t.Errorf("recentlyAdded = %d, want 2", len(recent))
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
