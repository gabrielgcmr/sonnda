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

	for _, path := range []string{"/", "/readyz"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200: %s", path, response.Code, response.Body.String())
		}
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
		"operationId: getApiMetadata",
		"operationId: getReadiness",
		"operationId: createCurrentAccount",
		"operationId: getCurrentAccount",
		"operationId: updateCurrentAccount",
		"operationId: deleteCurrentAccount",
		"operationId: listCurrentAccountPatients",
		"operationId: createPatient",
		"operationId: listPatients",
		"operationId: getPatient",
		"operationId: listExamDocuments",
		"operationId: uploadExamDocument",
		"operationId: getExamDocument",
		"operationId: getExamDocumentFile",
		"operationId: listExamDocumentTexts",
		"operationId: listPatientLabReports",
		"operationId: getLabReport",
		"bearerAuth:",
		"/me:",
		"/readyz:",
		"/me/patients:",
		"/patients:",
		"/patients/{patientId}:",
		"/patients/{patientId}/exam-documents:",
		"/exam-documents/{documentId}:",
		"/exam-documents/{documentId}/file:",
		"/patients/{patientId}/exam-document-texts:",
		"/patients/{patientId}/lab-reports:",
		"/lab-reports/{labReportId}:",
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

func TestLegacyLabsRouteIsNotRegistered(t *testing.T) {
	router := newAccountRouter(&accountUserRepository{})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/patients/019a1f08-29a2-7b47-929d-bdc50bb59919/labs", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("GET legacy labs route = %d, want 404", response.Code)
	}
}
