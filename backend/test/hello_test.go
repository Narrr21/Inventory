// test/hello_test.go
package test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	handler "my-backend/api"
)

func TestHandler_CORSHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rec := httptest.NewRecorder()

	handler.Handler(rec, req)

	res := rec.Result()

	tests := []struct {
		header   string
		expected string
	}{
		{"Access-Control-Allow-Origin", "*"},
		{"Access-Control-Allow-Methods", "GET, OPTIONS"},
		{"Access-Control-Allow-Headers", "Content-Type"},
		{"Content-Type", "application/json"},
	}

	for _, tt := range tests {
		if got := res.Header.Get(tt.header); got != tt.expected {
			t.Errorf("header %q = %q, want %q", tt.header, got, tt.expected)
		}
	}
}

func TestHandler_PreflightOptions(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/hello", nil)
	rec := httptest.NewRecorder()

	handler.Handler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status code untuk OPTIONS = %d, want %d", res.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(res.Body)
	if len(body) != 0 {
		t.Errorf("body untuk OPTIONS harus kosong, dapat: %q", body)
	}
}

func TestHandler_GetReturnsExpectedMessage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rec := httptest.NewRecorder()

	handler.Handler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want %d", res.StatusCode, http.StatusOK)
	}

	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("gagal decode JSON response: %v", err)
	}

	want := "Halo dari Backend Go di Vercel!"
	if body["message"] != want {
		t.Errorf("message = %q, want %q", body["message"], want)
	}
}