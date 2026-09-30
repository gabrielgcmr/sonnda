// internal/api/humaerror/huma_error_test.go
package humaerror_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/api/humaerror"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

func TestFromMapsApplicationErrorToHumaModel(t *testing.T) {
	model := humaerror.From(apperr.DomainRuleViolation(
		"data inválida",
		apperr.Violation{Field: "body.birth_date", Reason: "invalid_date"},
	))

	if model.Status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", model.Status, http.StatusUnprocessableEntity)
	}
	if model.Detail != "data inválida" {
		t.Fatalf("detail = %q, want %q", model.Detail, "data inválida")
	}
	if len(model.Errors) != 1 || model.Errors[0].Location != "body.birth_date" || model.Errors[0].Message != "invalid_date" {
		t.Fatalf("errors = %#v, want Huma detail for the violation", model.Errors)
	}
}

func TestFromHidesUnknownError(t *testing.T) {
	model := humaerror.From(errors.New("database connection refused"))

	if model.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", model.Status, http.StatusInternalServerError)
	}
	if model.Detail != "erro inesperado" {
		t.Fatalf("detail = %q, want safe message", model.Detail)
	}
}
