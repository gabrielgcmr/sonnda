// internal/app/config/config_test.go
package config

import (
	"errors"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv(envSupabaseURL, "postgres://user:pass@localhost:5432/db")
	t.Setenv(envSupabaseProjectURL, "https://project.supabase.co")
	t.Setenv(envGCPProjectID, "sonnda")
	t.Setenv(envGCPProjectNumber, "123456")
	t.Setenv(envGCSBucket, "sonnda-bucket")
	t.Setenv(envGCPLocation, "us")
	t.Setenv(envGCPExtractLabsProcessorID, "processor")
	t.Setenv(envRedisURL, "redis://localhost:6379")
	t.Setenv(envGoogleApplicationCredentials, "/tmp/sonnda-creds.json")
}

func TestLoadDefaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Env != "dev" {
		t.Fatalf("expected Env=dev, got %q", cfg.Env)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("expected LogLevel=info, got %q", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Fatalf("expected LogFormat=text, got %q", cfg.LogFormat)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected Port=8080, got %q", cfg.Port)
	}
	if cfg.DBURL == "" {
		t.Fatal("expected DBURL to be set, got empty")
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv(envLogLevel, "verbose")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if !hasViolation(appErr, envLogLevel) {
		t.Fatalf("expected violation for %s", envLogLevel)
	}
}

func TestLoadUsesSupabaseURLWhenProvided(t *testing.T) {
	setRequiredEnv(t)
	expected := "postgres://user:pass@localhost:5432/db"
	t.Setenv(envSupabaseURL, expected)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.DBURL != expected {
		t.Fatalf("expected DBURL=%q, got %q", expected, cfg.DBURL)
	}
}

func hasViolation(err *apperr.AppError, field string) bool {
	for _, v := range err.Violations {
		if v.Field == field {
			return true
		}
	}
	return false
}
