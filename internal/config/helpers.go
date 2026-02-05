// internal/config/helpers.go
package config

import (
	"os"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

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
