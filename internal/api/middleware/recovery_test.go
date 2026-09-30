// internal/api/middleware/recovery_test.go
package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gabrielgcmr/sonnda/internal/api/humaerror"
	"github.com/gabrielgcmr/sonnda/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

func TestRecoveryUsesHumaProblemAndLogsPanicOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, kind := range []string{"huma", "gin", "middleware", "committed"} {
		t.Run(kind, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			router := gin.New()
			var api huma.API
			router.Use(middleware.RequestID(), middleware.AccessLog(logger),
				middleware.Recovery(logger, func(c *gin.Context, err error) {
					op := &huma.Operation{Method: c.Request.Method, Path: c.FullPath()}
					if err := humaerror.Write(api, humagin.NewContext(op, c), err); err != nil {
						t.Error(err)
					}
				}))
			if kind == "middleware" {
				router.Use(func(c *gin.Context) { panic("private middleware panic") })
			}
			cfg := huma.DefaultConfig("test", "test")
			cfg.Transformers = append(cfg.Transformers, humaerror.Transform)
			api = humagin.New(router, cfg)
			if kind == "huma" {
				huma.Register(api, huma.Operation{
					OperationID: "panic", Method: http.MethodGet, Path: "/panic",
				}, func(context.Context, *struct{}) (*struct{}, error) {
					panic("private handler panic")
				})
			} else {
				router.GET("/panic", func(c *gin.Context) {
					if kind == "middleware" {
						t.Fatal("handler executed after middleware panic")
					}
					if kind == "committed" {
						c.String(http.StatusOK, "already sent")
					}
					panic("private handler panic")
				})
			}
			request := httptest.NewRequest(http.MethodGet, "/panic", nil)
			request.Header.Set("X-Request-ID", "recovery-test")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			wantStatus := http.StatusInternalServerError
			if kind == "committed" {
				wantStatus = http.StatusOK
				if response.Body.String() != "already sent" {
					t.Fatalf("committed response was overwritten: %s", response.Body.String())
				}
			} else {
				if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/problem+json") {
					t.Fatalf("unexpected content type: %s", response.Header().Get("Content-Type"))
				}
				var problem huma.ErrorModel
				if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
					t.Fatal(err)
				}
				if problem.Status != 500 || problem.Title != "Internal Server Error" || problem.Detail != "erro inesperado" {
					t.Fatalf("unexpected Huma problem: %+v", problem)
				}
				for _, private := range []string{"private", "stack", "error_code", "timestamp", "traceId"} {
					if strings.Contains(response.Body.String(), private) {
						t.Fatalf("unexpected private/legacy field: %s", response.Body.String())
					}
				}
			}
			if response.Code != wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, wantStatus)
			}
			accessCount, panicCount := 0, 0
			for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
				var entry map[string]any
				if err := json.Unmarshal([]byte(line), &entry); err != nil {
					t.Fatal(err)
				}
				if entry["request_id"] != "recovery-test" {
					t.Fatalf("missing request correlation: %s", line)
				}
				switch entry["msg"] {
				case "panic_recovered":
					panicCount++
					stack, _ := entry["stack"].(string)
					if stack == "" || !strings.Contains(line, "private") {
						t.Fatalf("missing panic diagnostics: %s", line)
					}
				case "handler_error":
					t.Fatalf("panic was logged twice: %s", logs.String())
				default:
					accessCount++
					if entry["error_code"] != "INTERNAL_ERROR" || entry["status"] != float64(wantStatus) {
						t.Fatalf("unexpected access metadata: %s", line)
					}
				}
			}
			if accessCount != 1 || panicCount != 1 {
				t.Fatalf("access=%d panic=%d, want one each: %s", accessCount, panicCount, logs.String())
			}
		})
	}
}
