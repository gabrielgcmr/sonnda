// internal/kernel/apperr/predicates_test.go
package apperr

import (
	"errors"
	"testing"
)

func TestHasCode(t *testing.T) {
	err := &AppError{Kind: NOT_FOUND, Message: "x"}

	if !HasCode(err, NOT_FOUND) {
		t.Fatalf("expected HasCode to be true")
	}
	if HasCode(err, VALIDATION_FAILED) {
		t.Fatalf("expected HasCode to be false")
	}
	if HasCode(nil, NOT_FOUND) {
		t.Fatalf("expected HasCode(nil, ...) to be false")
	}
	if HasCode(err) {
		t.Fatalf("expected HasCode(err) to be false")
	}
}

func TestPredicatesWithWrappedError(t *testing.T) {
	base := &AppError{Kind: VALIDATION_FAILED, Message: "x"}
	wrapped := errors.Join(errors.New("other"), base)

	if !IsValidation(wrapped) {
		t.Fatalf("expected IsValidation to be true for wrapped AppError")
	}
	if IsNotFound(wrapped) {
		t.Fatalf("expected IsNotFound to be false")
	}
}
