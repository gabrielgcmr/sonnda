// internal/api/app_test.go
package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gin-gonic/gin"
)

func TestAppTracesCORSResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const allowedOrigin = "https://app.example.test"
	for _, tc := range []struct {
		name, method, origin, path string
		status                     int
		allowOrigin                string
	}{
		{"preflight before authentication", http.MethodOptions, allowedOrigin, "/me", http.StatusNoContent, allowedOrigin},
		{"rejected origin", http.MethodGet, "https://blocked.example.test", "/healthz", http.StatusForbidden, ""},
		{"allowed origin", http.MethodGet, allowedOrigin, "/healthz", http.StatusOK, allowedOrigin},
		{"allowed origin with error", http.MethodGet, allowedOrigin, "/me", http.StatusUnauthorized, allowedOrigin},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			app := New(Options{
				Deps:   &APIDependencies{},
				Logger: slog.New(slog.NewJSONHandler(&logs, nil)),
				CORSConfig: config.CORSConfig{
					AllowOrigins:  []string{allowedOrigin},
					AllowMethods:  []string{"GET", "OPTIONS"},
					AllowHeaders:  []string{"Authorization", "X-Request-ID"},
					ExposeHeaders: []string{"X-Request-ID"},
				},
			})
			request := httptest.NewRequest(tc.method, tc.path, nil)
			request.Header.Set("Origin", tc.origin)
			if tc.method == http.MethodOptions {
				request.Header.Set("Access-Control-Request-Method", "GET")
				request.Header.Set("Access-Control-Request-Headers", "Authorization")
			}
			response := httptest.NewRecorder()
			app.router.ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tc.status, response.Body.String())
			}
			if got := response.Header().Get("Access-Control-Allow-Origin"); got != tc.allowOrigin {
				t.Fatalf("allow origin = %q, want %q", got, tc.allowOrigin)
			}
			requestID := response.Header().Get("X-Request-ID")
			if requestID == "" {
				t.Fatal("missing request ID on response")
			}
			lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
			if len(lines) != 1 {
				t.Fatalf("expected one access log: %s", logs.String())
			}
			var entry map[string]any
			if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
				t.Fatal(err)
			}
			if entry["request_id"] != requestID || entry["status"] != float64(tc.status) ||
				entry["method"] != tc.method || entry["path"] != tc.path {
				t.Fatalf("unexpected access metadata: %s", lines[0])
			}
		})
	}
}
