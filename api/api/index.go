// Package handler contém a entrada experimental da Vercel. Não habilitar em
// produção até a fábrica possuir execução e arquivos persistentes.
package handler

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/geevheit/neuralcenter/api/config"
	"github.com/geevheit/neuralcenter/api/internal/bootstrap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	once       sync.Once
	appHandler http.Handler
)

// Handler é a entrada HTTP do runtime Go da Vercel.
func Handler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("GEEV_SERVERLESS_EXPERIMENTAL") != "1" {
		http.Error(w, "API serverless indisponível durante a migração", http.StatusServiceUnavailable)
		return
	}
	once.Do(func() {
		if os.Getenv("DATABASE_URL") == "" || os.Getenv("SUPABASE_URL") == "" || os.Getenv("OWNER_USER_ID") == "" {
			return
		}
		cfg := config.Load()
		poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
		if err != nil {
			return
		}
		if cfg.DBSimpleProtocol {
			poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		}
		// Limita conexões por instância serverless; o pool é compartilhado na instância quente.
		poolConfig.MaxConns = 2
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			return
		}
		appHandler = bootstrap.NewHandler(cfg, pool, slog.Default())
	})
	if appHandler == nil {
		http.Error(w, "Configuração da API serverless indisponível", http.StatusServiceUnavailable)
		return
	}
	appHandler.ServeHTTP(w, r)
}
