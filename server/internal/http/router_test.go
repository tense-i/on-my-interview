package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterHealth(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewRouter(Dependencies{Version: "v0.0.1"}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", payload["status"])
	}
	if payload["version"] != "v0.0.1" {
		t.Fatalf("expected version v0.0.1, got %q", payload["version"])
	}
}
