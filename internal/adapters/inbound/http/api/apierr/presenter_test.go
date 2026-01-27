package apierr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/gabrielgcmr/sonnda/internal/app/apperr"
)

func TestWriteError_IncludesViolations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/register", nil)

	ErrorResponder(c, &apperr.AppError{
		Code:    apperr.VALIDATION_FAILED,
		Message: "validacao falhou",
		Violations: []apperr.Violation{
			{Field: "full_name", Reason: "required"},
		},
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	errObj, ok := raw["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'error' object in response, got %T", raw["error"])
	}

	if errObj["code"] != string(apperr.VALIDATION_FAILED) {
		t.Fatalf("expected code %q, got %v", apperr.VALIDATION_FAILED, errObj["code"])
	}

	violations, ok := errObj["violations"].([]any)
	if !ok {
		t.Fatalf("expected 'violations' array in response, got %T", errObj["violations"])
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v0, ok := violations[0].(map[string]any)
	if !ok {
		t.Fatalf("expected violation to be object, got %T", violations[0])
	}

	if v0["field"] != "full_name" {
		t.Fatalf("expected violation field %q, got %v", "full_name", v0["field"])
	}
	if v0["reason"] != "required" {
		t.Fatalf("expected violation reason %q, got %v", "required", v0["reason"])
	}
}

func TestWriteError_OmitsViolationsWhenEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/register", nil)

	ErrorResponder(c, &apperr.AppError{
		Code:    apperr.VALIDATION_FAILED,
		Message: "payload invalido",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	errObj, ok := raw["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'error' object in response, got %T", raw["error"])
	}

	if _, ok := errObj["violations"]; ok {
		t.Fatalf("did not expect 'violations' in response when empty")
	}
}
