// internal/app/config/config.go
package config

import (
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

const (
	envGoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	envGCPProjectID                 = "GCP_PROJECT_ID"
	envGCPProjectNumber             = "GCP_PROJECT_NUMBER"
	envGCSBucket                    = "GCS_BUCKET"
	envGCPLocation                  = "GCP_LOCATION"
	envGCPExtractLabsProcessorID    = "GCP_EXTRACT_LABS_PROCESSOR_ID"

	envSupabaseURL        = "SUPABASE_URL"
	envSupabaseProjectURL = "SUPABASE_PROJECT_URL"
	envSupabaseJWTIssuer  = "SUPABASE_JWT_ISSUER"
	envSupabaseJWTAud     = "SUPABASE_JWT_AUDIENCE"
	envRedisURL           = "REDIS_URL"

	envAPIHost   = "API_HOST"
	envPort      = "PORT"
	envAppEnv    = "APP_ENV"
	envLogLevel  = "LOG_LEVEL"
	envLogFormat = "LOG_FORMAT"
)

var (
	allowedEnvs       = map[string]struct{}{"dev": {}, "prod": {}}
	allowedLogLevels  = map[string]struct{}{"debug": {}, "info": {}, "warn": {}, "warning": {}, "error": {}}
	allowedLogFormats = map[string]struct{}{"text": {}, "json": {}, "pretty": {}}

	errInvalidHost = errors.New("invalid host")
)

type Config struct {
	DBURL    string
	RedisURL string

	GoogleApplicationCredentials string
	GCPProjectID                 string
	GCPProjectNumber             string
	GCSBucket                    string
	GCPLocation                  string
	GCPExtractLabsProcessorID    string
	SupabaseProjectURL           string
	SupabaseJWTIssuer            string
	SupabaseJWTAudience          string

	APIHost string

	Port string // porta HTTP (ex.: "8080")
	Env  string // ex.: "dev", "prod"

	LogLevel  string
	LogFormat string
}

func Load() (*Config, error) {
	// Carrega variáveis do arquivo .env
	_ = godotenv.Load()

	cfg := &Config{
		DBURL:    getEnv(envSupabaseURL),
		RedisURL: getEnv(envRedisURL),
		// Google
		GoogleApplicationCredentials: getEnv(envGoogleApplicationCredentials),
		GCPProjectID:                 getEnv(envGCPProjectID),
		GCPProjectNumber:             getEnv(envGCPProjectNumber),
		GCSBucket:                    getEnv(envGCSBucket),
		GCPLocation:                  getEnv(envGCPLocation),
		GCPExtractLabsProcessorID:    getEnv(envGCPExtractLabsProcessorID),
		SupabaseProjectURL:           getEnv(envSupabaseProjectURL),
		SupabaseJWTIssuer:            getEnv(envSupabaseJWTIssuer),
		SupabaseJWTAudience:          getEnv(envSupabaseJWTAud),

		Port:      getEnvOrDefault(envPort, "8080"),
		Env:       strings.ToLower(getEnvOrDefault(envAppEnv, "dev")),
		LogLevel:  strings.ToLower(getEnvOrDefault(envLogLevel, "info")),
		LogFormat: strings.ToLower(getEnvOrDefault(envLogFormat, "text")),
	}

	rawAPIHost := getEnv(envAPIHost)
	if rawAPIHost == "" {
		switch cfg.Env {
		case "dev":
			rawAPIHost = "api.localhost"
		case "prod":
			rawAPIHost = "api.sonnda.com.br"
		}
	}

	var violations []apperr.Violation

	if host, err := normalizeHost(rawAPIHost); err != nil {
		violations = append(violations, apperr.Violation{
			Field:  envAPIHost,
			Reason: "invalid_host",
		})
	} else {
		cfg.APIHost = host
	}

	appendRequired(&violations, envSupabaseURL, cfg.DBURL)
	appendRequired(&violations, envSupabaseProjectURL, cfg.SupabaseProjectURL)
	appendRequired(&violations, envGCPProjectID, cfg.GCPProjectID)
	appendRequired(&violations, envGCSBucket, cfg.GCSBucket)
	appendRequired(&violations, envGCPLocation, cfg.GCPLocation)
	appendRequired(&violations, envGCPExtractLabsProcessorID, cfg.GCPExtractLabsProcessorID)
	appendRequired(&violations, envRedisURL, cfg.RedisURL)
	appendRequired(&violations, envGoogleApplicationCredentials, cfg.GoogleApplicationCredentials)
	validateEnum(&violations, envAppEnv, cfg.Env, allowedEnvs)
	validateEnum(&violations, envLogLevel, cfg.LogLevel, allowedLogLevels)
	validateEnum(&violations, envLogFormat, cfg.LogFormat, allowedLogFormats)

	if len(violations) > 0 {
		return nil, apperr.Validation("invalid configuration", violations...)
	}

	if cfg.Env == "prod" && cfg.LogFormat == "text" {
		cfg.LogFormat = "json"
	}

	return cfg, nil
}

func getEnvOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func appendRequired(violations *[]apperr.Violation, field, value string) {
	if strings.TrimSpace(value) == "" {
		*violations = append(*violations, apperr.Violation{
			Field:  field,
			Reason: "required",
		})
	}
}

func validateEnum(violations *[]apperr.Violation, field, value string, allowed map[string]struct{}) {
	if value == "" {
		return
	}
	if _, ok := allowed[strings.ToLower(value)]; !ok {
		*violations = append(*violations, apperr.Violation{
			Field:  field,
			Reason: "invalid_enum",
		})
	}
}

func normalizeHost(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	var u *url.URL
	var err error
	if strings.Contains(raw, "://") {
		u, err = url.Parse(raw)
	} else {
		u, err = url.Parse("//" + raw)
	}
	if err != nil {
		return "", errInvalidHost
	}

	if u.Hostname() == "" {
		return "", errInvalidHost
	}
	if u.Path != "" && u.Path != "/" {
		return "", errInvalidHost
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", errInvalidHost
	}

	return strings.ToLower(u.Hostname()), nil
}
