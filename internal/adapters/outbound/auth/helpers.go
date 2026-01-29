// internal/adapters/outbound/auth/helpers.go
package auth

import (
	"os"
	"strings"
)

func stringPtr(v string) *string {
	return &v
}

func envOrFallback(key string, fallbacks ...string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	for _, fb := range fallbacks {
		if v := strings.TrimSpace(os.Getenv(fb)); v != "" {
			return v
		}
	}
	return ""
}
