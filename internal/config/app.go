// internal/config/app.go
package config

import "strings"

const (
	envAppEnv    = "APP_ENV"
	envLogLevel  = "LOG_LEVEL"
	envLogFormat = "LOG_FORMAT"
)

var (
	allowedEnvs       = map[string]struct{}{"dev": {}, "prod": {}}
	allowedLogLevels  = map[string]struct{}{"debug": {}, "info": {}, "warn": {}, "warning": {}, "error": {}}
	allowedLogFormats = map[string]struct{}{"text": {}, "json": {}, "pretty": {}}
)

type AppConfig struct {
	Env       string
	LogLevel  string
	LogFormat string
}

func loadAppConfig() AppConfig {
	return AppConfig{
		Env:       strings.ToLower(getEnvOrDefault(envAppEnv, "dev")),
		LogLevel:  strings.ToLower(getEnvOrDefault(envLogLevel, "info")),
		LogFormat: strings.ToLower(getEnvOrDefault(envLogFormat, "text")),
	}
}
