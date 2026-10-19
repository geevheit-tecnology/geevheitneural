package config

import (
	"bufio"
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
	loadDotEnv(".env")
	loadDotEnv("../.env")

	port := getEnv("PORT", "8085")
	logLevel := getEnv("LOG_LEVEL", "info")
	dbURL := getEnv("DATABASE_URL", "postgresql://postgres:postgres@127.0.0.1:54332/postgres?sslmode=disable")
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

func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}

