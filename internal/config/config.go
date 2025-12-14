// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"sonnda-api/internal/adapters/outbound/database/supabase"
	"strings"
	"time"
)

type Config struct {
	DBURL string

	GCPProjectID    string
	GCSBucket       string
	GCPLocation     string
	LabsProcessorID string

	JWTSecret string

	Port string // porta HTTP (ex.: "8080")
	Env  string // ex.: "dev", "prod"

	LogLevel  string
	LogFormat string
}

func Load() (*Config, error) {
	cfg := &Config{
		DBURL:           os.Getenv("DATABASE_URL"),
		GCPProjectID:    os.Getenv("GCP_PROJECT_ID"),
		GCSBucket:       os.Getenv("GCS_BUCKET"),
		GCPLocation:     os.Getenv("GCP_LOCATION"),
		LabsProcessorID: os.Getenv("DOCAI_LABS_PROCESSOR_ID"),
		JWTSecret:       os.Getenv("SUPABASE_JWT_SECRET"),
		Port:            getEnvOrDefault("PORT", "8080"),
		Env:             getEnvOrDefault("APP_ENV", "dev"),
		LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
		LogFormat:       getEnvOrDefault("LOG_FORMAT", "text"),
	}

	if cfg.Env == "prod" && cfg.LogFormat == "text" {
		cfg.LogFormat = "json"
	}

	// validação básica dos obrigatórios
	var missing []string

	if cfg.DBURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.GCPProjectID == "" {
		missing = append(missing, "GCP_PROJECT_ID")
	}
	if cfg.GCSBucket == "" {
		missing = append(missing, "GCS_BUCKET")
	}
	if cfg.GCPLocation == "" {
		missing = append(missing, "GCP_LOCATION")
	}
	if cfg.LabsProcessorID == "" {
		missing = append(missing, "DOCAI_LABS_PROCESSOR_ID")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func SupabaseConfig(cfg Config) supabase.Config {
	return supabase.Config{
		DatabaseURL:     cfg.DBURL,
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}
}
