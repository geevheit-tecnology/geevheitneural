// Package serverless expõe a montagem da API para o adaptador do runtime Vercel.
package serverless

import (
	"log/slog"
	"net/http"

	"github.com/geevheit/neuralcenter/api/config"
	"github.com/geevheit/neuralcenter/api/internal/bootstrap"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewHandler(cfg *config.Config, pool *pgxpool.Pool, logger *slog.Logger) http.Handler {
	return bootstrap.NewHandler(cfg, pool, logger)
}
