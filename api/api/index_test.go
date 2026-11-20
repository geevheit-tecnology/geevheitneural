package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerlessDisabledByDefault(t *testing.T) {
	t.Setenv("GEEV_SERVERLESS_EXPERIMENTAL", "")
	w := httptest.NewRecorder()
	Handler(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestServerlessRequiresExplicitConfiguration(t *testing.T) {
	t.Setenv("GEEV_SERVERLESS_EXPERIMENTAL", "1")
	t.Setenv("DATABASE_URL", "")
	w := httptest.NewRecorder()
	Handler(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
