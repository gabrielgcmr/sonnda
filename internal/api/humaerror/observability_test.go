// internal/api/humaerror/observability_test.go
package humaerror_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gabrielgcmr/sonnda/internal/api/humaerror"
	"github.com/gabrielgcmr/sonnda/internal/api/middleware"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
)

func TestFromPreservesCauseWithoutSerializingIt(t *testing.T) {
	cause := errors.New("private connection detail")
	appErr := apperr.Internal("serviço indisponível", cause)
	original := fmt.Errorf("operation failed: %w", appErr)
	converted := humaerror.From(original)
	if !errors.Is(converted, cause) || errors.Unwrap(converted) != original {
		t.Fatal("original error chain was lost")
	}
	var recovered *apperr.AppError
	if !errors.As(converted, &recovered) || recovered != appErr {
		t.Fatal("application error was lost")
	}
	body, err := json.Marshal(converted)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "private") || strings.Contains(string(body), "operation failed") {
		t.Fatalf("internal cause leaked: %s", body)
	}
}

func TestHumaErrorObservability(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, path := range []string{"handler", "middleware"} {
		for _, tc := range []struct {
			name   string
			err    error
			status int
			code   string
			causes []string
		}{
			{"server", apperr.Internal("serviço indisponível", errors.Join(
				fmt.Errorf("repository: %w", errors.New("private database cause")),
				errors.New("private cleanup cause"),
			)), 500, "INTERNAL_ERROR", []string{"private database cause", "private cleanup cause"}},
			{"unknown", errors.New("private unknown cause"), 500, "INTERNAL_ERROR", []string{"private unknown cause"}},
			{"client", apperr.Conflict("recurso já existe"), 409, "RESOURCE_CONFLICT", nil},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				var logs bytes.Buffer
				logger := slog.New(slog.NewJSONHandler(&logs, nil))
				router := gin.New()
				router.Use(middleware.RequestID(), middleware.AccessLog(logger))
				config := huma.DefaultConfig("test", "test")
				config.Transformers = append(config.Transformers, humaerror.Transform)
				api := humagin.New(router, config)
				if path == "middleware" {
					api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
						if err := humaerror.Write(api, ctx, tc.err); err != nil {
							t.Error(err)
						}
					})
				}
				huma.Register(api, huma.Operation{OperationID: "failure", Method: http.MethodGet, Path: "/failure"},
					func(context.Context, *struct{}) (*struct{}, error) {
						return nil, humaerror.From(tc.err)
					})
				request := httptest.NewRequest(http.MethodGet, "/failure", nil)
				request.Header.Set("X-Request-ID", "observability-test")
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				if response.Code != tc.status || !strings.HasPrefix(response.Header().Get("Content-Type"), "application/problem+json") {
					t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
				}
				if strings.Contains(response.Body.String(), "private") {
					t.Fatalf("private cause exposed: %s", response.Body.String())
				}
				var body map[string]any
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				for _, key := range []string{"cause", "code", "error_code", "error_chain"} {
					if _, exists := body[key]; exists {
						t.Fatalf("internal field %q in public model", key)
					}
				}
				accessCount, detailCount := 0, 0
				for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
					var entry map[string]any
					if err := json.Unmarshal([]byte(line), &entry); err != nil {
						t.Fatal(err)
					}
					if entry["request_id"] != "observability-test" || entry["error_code"] != tc.code || entry["status"] != float64(tc.status) {
						t.Fatalf("missing request/error metadata: %s", line)
					}
					if entry["msg"] == "handler_error" {
						detailCount++
						for _, cause := range tc.causes {
							if !strings.Contains(line, cause) {
								t.Fatalf("missing cause %q in detailed log: %s", cause, line)
							}
						}
					} else {
						accessCount++
						if strings.Contains(line, "private") {
							t.Fatal("access log must not contain detailed causes")
						}
					}
				}
				wantDetails := 0
				if tc.status >= 500 {
					wantDetails = 1
				}
				if accessCount != 1 || detailCount != wantDetails {
					t.Fatalf("access=%d details=%d, want 1/%d: %s", accessCount, detailCount, wantDetails, logs.String())
				}
			})
		}
	}
}
