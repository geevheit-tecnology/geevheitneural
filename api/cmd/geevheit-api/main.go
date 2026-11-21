package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/geevheit/neuralcenter/api/config"
	"github.com/geevheit/neuralcenter/api/internal/bootstrap"
)

func main() {
	// 1. Configuração e Logger estruturado (slog JSON sem PII)
	cfg := config.Load()

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("starting_geevheit_api",
		slog.String("port", cfg.Port),
		slog.String("log_level", cfg.LogLevel),
	)

	// 2. Conexão ao PostgreSQL via pgxpool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed_to_parse_database_url", slog.Any("error", err))
		os.Exit(1)
	}

	// Se configurado para Simple Protocol (Render / pooler Supabase porta 6543)
	if cfg.DBSimpleProtocol {
		poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("failed_to_connect_to_postgres", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Warn("database_ping_warning", slog.Any("error", err))
	} else {
		logger.Info("connected_to_postgres_successfully")
	}

	handler := bootstrap.NewHandler(cfg, pool, logger)

	// 5. Servidor HTTP com timeouts e Graceful Shutdown
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server_listening", slog.String("addr", server.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	// Captura sinais de interrupção do sistema operacional
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server_fatal_error", slog.Any("error", err))
		}

	case sig := <-shutdown:
		logger.Info("shutdown_signal_received", slog.String("signal", sig.String()))

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful_shutdown_failed", slog.Any("error", err))
			_ = server.Close()
		}
		logger.Info("server_stopped_gracefully")
	}
}
