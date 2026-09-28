// internal/api/huma_test.go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHumaServesHealthDocsAndOneOpenAPISpec(t *testing.T) {
	router := newAccountRouter(&accountUserRepository{})

	health := httptest.NewRecorder()
	router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	var body struct {
		Status string `json:"status"`
	}
	if health.Code != http.StatusOK || json.Unmarshal(health.Body.Bytes(), &body) != nil || body.Status != "ok" {
		t.Fatalf("GET /healthz = %d %s, want 200 with status ok", health.Code, health.Body.String())
	}

	for _, path := range []string{"/openapi.json", "/openapi.yaml", "/docs"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200: %s", path, response.Code, response.Body.String())
		}
	}

	spec := httptest.NewRecorder()
	router.ServeHTTP(spec, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	generated := spec.Body.String()
	for _, expected := range []string{
		"operationId: getHealth",
		"operationId: createCurrentAccount",
		"operationId: getCurrentAccount",
		"operationId: updateCurrentAccount",
		"operationId: deleteCurrentAccount",
		"operationId: listCurrentAccountPatients",
		"operationId: createPatient",
		"operationId: listPatients",
		"operationId: getPatient",
		"bearerAuth:",
		"/me:",
		"/me/patients:",
		"/patients:",
		"/patients/{patientId}:",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated OpenAPI is missing %q", expected)
		}
	}

	legacy := httptest.NewRecorder()
	router.ServeHTTP(legacy, httptest.NewRequest(http.MethodGet, "/v1/me", nil))
	if legacy.Code != http.StatusNotFound {
		t.Fatalf("GET /v1/me = %d, want 404", legacy.Code)
	}
}
