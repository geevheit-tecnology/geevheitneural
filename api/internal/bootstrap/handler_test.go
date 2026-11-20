package bootstrap

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlerHealthAndProtectedRoute(t *testing.T) {
	poolConfig, err := pgxpool.ParseConfig("postgres://localhost:1/test?connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	cfg := &config.Config{OwnerUserID: "owner", SupabaseURL: "https://example.invalid", APIBaseURL: "http://localhost", BaseURL: "http://localhost"}
	h := NewHandler(cfg, pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, tt := range []struct {
		path string
		want int
	}{{"/healthz", http.StatusOK}, {"/v1/agents", http.StatusUnauthorized}} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if w.Code != tt.want {
			t.Errorf("%s: got %d, want %d", tt.path, w.Code, tt.want)
		}
	}
}
