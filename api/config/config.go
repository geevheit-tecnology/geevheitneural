package config

import (
	"os"
	"strings"
)

type Config struct {
	Port               string
	LogLevel           string
	DatabaseURL        string
	DBSimpleProtocol   bool
	OwnerUserID        string
	SupabaseURL        string
	SupabaseJWTSecret  string
	CORSAllowedOrigins []string
	BaseURL            string
}

func Load() *Config {
	port := getEnv("PORT", "8085")
	logLevel := getEnv("LOG_LEVEL", "info")
	dbURL := getEnv("DATABASE_URL", "postgresql://geev_api:troque-esta-senha-local@127.0.0.1:54332/postgres?sslmode=disable")
	dbSimple := getEnv("DB_SIMPLE_PROTOCOL", "false") == "true"
	ownerID := getEnv("OWNER_USER_ID", "a0000000-0000-0000-0000-000000000001")
	sbURL := getEnv("SUPABASE_URL", "http://127.0.0.1:54331")
	sbSecret := getEnv("SUPABASE_JWT_SECRET", "super-secret-jwt-token-with-at-least-32-bytes-long")
	corsStr := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,http://127.0.0.1:3000")
	baseURL := getEnv("BASE_URL", "http://localhost:3000")

	var origins []string
	for _, o := range strings.Split(corsStr, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return &Config{
		Port:               port,
		LogLevel:           logLevel,
		DatabaseURL:        dbURL,
		DBSimpleProtocol:   dbSimple,
		OwnerUserID:        ownerID,
		SupabaseURL:        sbURL,
		SupabaseJWTSecret:  sbSecret,
		CORSAllowedOrigins: origins,
		BaseURL:            baseURL,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
