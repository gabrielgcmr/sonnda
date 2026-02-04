// internal/api/routes_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gin-gonic/gin"
)

func TestRootRoute_ReturnsMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerRootRoute(r, &config.Config{Env: "prod"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var got RootResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if got.Name != rootAPIName {
		t.Fatalf("expected name %q, got %q", rootAPIName, got.Name)
	}
	if got.Version != rootAPIVersion {
		t.Fatalf("expected version %q, got %q", rootAPIVersion, got.Version)
	}
	if got.Environment != "prod" {
		t.Fatalf("expected environment %q, got %q", "prod", got.Environment)
	}
	if got.Docs != "/docs" {
		t.Fatalf("expected docs %q, got %q", "/docs", got.Docs)
	}
	if got.OpenAPI != "/openapi.yaml" {
		t.Fatalf("expected openapi %q, got %q", "/openapi.yaml", got.OpenAPI)
	}
	if got.Health != "/healthz" {
		t.Fatalf("expected health %q, got %q", "/healthz", got.Health)
	}
	if got.Ready != "/readyz" {
		t.Fatalf("expected ready %q, got %q", "/readyz", got.Ready)
	}
}

func TestOpenAPIRoute_ReturnsSpec(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerOpenAPIRoute(r)

	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if body := resp.Body.Bytes(); len(body) == 0 {
		t.Fatal("expected non-empty spec body")
	}

	contentType := resp.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/yaml") {
		t.Fatalf("expected yaml content-type, got %q", contentType)
	}
}

func TestHealthzRoute_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerPublicRoutes(r)

	assertStatusOK(t, r, "/healthz")
}

func TestReadyzRoute_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerPublicRoutes(r)

	assertStatusOK(t, r, "/readyz")
}

func TestDocsRoute_ReturnsHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	registerDocsRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	contentType := resp.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected html content-type, got %q", contentType)
	}
	if !strings.Contains(resp.Body.String(), "redoc") {
		t.Fatal("expected docs html to include redoc tag or script")
	}
}

func assertStatusOK(t *testing.T, r *gin.Engine, path string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var got map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got["status"] != "ok" {
		t.Fatalf("expected status %q, got %q", "ok", got["status"])
	}
}
