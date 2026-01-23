// internal/app/config/config.go
package config

import (
	"fmt"
	"os"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	"strings"
	"time"
)

type Config struct {
	DBURL string

	GCPProjectID     string
	GCPProjectNumber string
	GCSBucket        string
	GCPLocation      string
	LabsProcessorID  string

	FirebaseProjectID         string
	FirebaseCredentialsFile   string
	FirebaseAPIKey            string
	FirebaseAuthDomain        string
	FirebaseStorageBucket     string
	FirebaseMessagingSenderID string
	FirebaseAppID             string

	AppHost string
	APIHost string

	Port string // porta HTTP (ex.: "8080")
	Env  string // ex.: "dev", "prod"

	LogLevel  string
	LogFormat string
}

func Load() (*Config, error) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	projectNumber := os.Getenv("GCP_PROJECT_NUMBER")
	gcsBucket := os.Getenv("GCS_BUCKET")

	cfg := &Config{
		DBURL:                     os.Getenv("SUPABASE_URL"),
		GCPProjectID:              projectID,
		GCPProjectNumber:          projectNumber,
		GCSBucket:                 gcsBucket,
		GCPLocation:               os.Getenv("GCP_LOCATION"),
		LabsProcessorID:           os.Getenv("DOCAI_LABS_PROCESSOR_ID"),
		FirebaseProjectID:         projectID,
		FirebaseCredentialsFile:   os.Getenv("FIREBASE_CREDENTIALS_FILE"),
		FirebaseAPIKey:            os.Getenv("FIREBASE_API_KEY"),
		FirebaseAuthDomain:        os.Getenv("FIREBASE_AUTH_DOMAIN"),
		FirebaseStorageBucket:     gcsBucket,
		FirebaseMessagingSenderID: projectNumber,
		FirebaseAppID:             os.Getenv("FIREBASE_APP_ID"),
		AppHost:                   normalizeHost(os.Getenv("APP_HOST")),
		APIHost:                   normalizeHost(os.Getenv("API_HOST")),
		Port:                      getEnvOrDefault("PORT", "8080"),
		Env:                       getEnvOrDefault("APP_ENV", "dev"),
		LogLevel:                  getEnvOrDefault("LOG_LEVEL", "info"),
		LogFormat:                 getEnvOrDefault("LOG_FORMAT", "text"),
	}

	if cfg.Env == "prod" && cfg.LogFormat == "text" {
		cfg.LogFormat = "json"
	}

	// validação básica dos obrigatórios
	var missing []string

	if cfg.DBURL == "" {
		missing = append(missing, "SUPABASE_URL")
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

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return ""
	}
	return strings.Split(host, ":")[0]
}

func SupabaseConfig(cfg Config) db.Config {
	return db.Config{
		DatabaseURL:     cfg.DBURL,
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}
}
