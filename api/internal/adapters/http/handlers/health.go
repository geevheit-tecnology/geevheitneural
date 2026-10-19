package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HandleHealthz(w http.ResponseWriter, _ *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func HandleReadyz(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			middleware.WriteError(w, http.StatusServiceUnavailable, "database_unavailable", "banco de dados inacessível")
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ready",
		})
	}
}
