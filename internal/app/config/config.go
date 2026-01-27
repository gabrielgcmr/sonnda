// internal/app/config/config.go
package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres/repository/db"
)

type Config struct {
	DBURL string

	GCPProjectID     string
	GCPProjectNumber string
	GCSBucket        string
	GCPLocation      string
	LabsProcessorID  string

	FirebaseProjectID         string
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
		DBURL:             os.Getenv("SUPABASE_URL"),
		GCPProjectID:      projectID,
		GCPProjectNumber:  projectNumber,
		GCSBucket:         gcsBucket,
		GCPLocation:       os.Getenv("GCP_LOCATION"),
		LabsProcessorID:   os.Getenv("DOCAI_LABS_PROCESSOR_ID"),
		FirebaseProjectID: projectID,

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

// SetupGoogleCredentials configura GOOGLE_APPLICATION_CREDENTIALS
// Tenta usar arquivo local em dev, ou decodifica base64 em produção
func SetupGoogleCredentials() error {
	// Se já está setado, valida se o arquivo existe
	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		if _, err := os.Stat(credPath); err == nil {
			return nil // Arquivo existe, tudo certo
		}
		// Arquivo não existe, continua para tentar base64
	}

	// Tenta usar arquivo local primeiro (desenvolvimento)
	localPath := "secrets/sonnda-gcs.json"
	if _, err := os.Stat(localPath); err == nil {
		return os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", localPath)
	}

	// Em produção, tenta decodificar base64
	credB64 := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_B64")
	if credB64 == "" {
		return errors.New("GOOGLE_APPLICATION_CREDENTIALS_B64 não definido em produção")
	}

	// Decodifica o base64
	credJSON, err := base64.StdEncoding.DecodeString(credB64)
	if err != nil {
		return errors.New("erro ao decodificar GOOGLE_APPLICATION_CREDENTIALS_B64: " + err.Error())
	}

	// Escreve em arquivo temporário
	tmpDir := os.TempDir()
	credPath := filepath.Join(tmpDir, "gcp-credentials.json")
	if err := os.WriteFile(credPath, credJSON, 0600); err != nil {
		return errors.New("erro ao escrever credencial temporária: " + err.Error())
	}

	// Seta a variável de ambiente
	return os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credPath)
}
