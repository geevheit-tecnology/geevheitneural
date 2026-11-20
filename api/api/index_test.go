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

func TestForwardVercelRoute(t *testing.T) {
	tests := []struct {
		input, wantPath, wantQuery string
		ok                         bool
	}{
		{"/api/index?__geev_route=v1/cards/123&status=active", "/v1/cards/123", "status=active", true},
		{"/v1/cards?__geev_capture=cards&__geev_route=v1/cards&path=original", "/v1/cards", "path=original", true},
		{"/api/index?__geev_route=healthz", "/healthz", "", true},
		{"/v1/cards?status=active", "/v1/cards", "status=active", true},
		{"/api/index?__geev_route=v1/../healthz", "", "", false},
		{"/api/index?__geev_route=v1/cards&__geev_route=healthz", "", "", false},
		{"/api/index?__geev_route=unexpected", "", "", false},
		{"/api/index", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.input, nil)
			got, ok := forwardVercelRoute(req)
			if ok != tt.ok {
				t.Fatalf("ok = %v", ok)
			}
			if ok && (got.URL.Path != tt.wantPath || got.URL.RawQuery != tt.wantQuery) {
				t.Fatalf("got %s?%s", got.URL.Path, got.URL.RawQuery)
			}
		})
	}
}
