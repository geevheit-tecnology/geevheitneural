// Package handler contém a entrada experimental da Vercel. Não habilitar em
// produção até a fábrica possuir execução e arquivos persistentes.
package handler

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/geevheit/neuralcenter/api/config"
	"github.com/geevheit/neuralcenter/api/serverless"
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
		appHandler = serverless.NewHandler(cfg, pool, slog.Default())
	})
	if appHandler == nil {
		http.Error(w, "Configuração da API serverless indisponível", http.StatusServiceUnavailable)
		return
	}
	forwarded, ok := forwardVercelRoute(r)
	if !ok {
		http.Error(w, "Rota inválida", http.StatusBadRequest)
		return
	}
	appHandler.ServeHTTP(w, forwarded)
}

// forwardVercelRoute restaura o caminho capturado pelo rewrite quando o runtime
// entrega /api/index. A autenticação continua sendo aplicada pelo handler da API.
func forwardVercelRoute(r *http.Request) (*http.Request, bool) {
	query := r.URL.Query()
	values := query["__geev_route"]
	query.Del("__geev_route")
	query.Del("__geev_capture")
	copyRequest := r.Clone(r.Context())
	copyURL := *r.URL
	copyURL.RawQuery = query.Encode()
	copyRequest.URL = &copyURL
	if r.URL.Path != "/api/index" {
		return copyRequest, true
	}
	if len(values) != 1 || strings.ContainsAny(values[0], "\\?#\x00") {
		return nil, false
	}
	target := "/" + values[0]
	if path.Clean(target) != target || !(target == "/healthz" || target == "/readyz" || strings.HasPrefix(target, "/v1/")) {
		return nil, false
	}
	copyURL.Path = target
	copyURL.RawPath = ""
	return copyRequest, true
}
