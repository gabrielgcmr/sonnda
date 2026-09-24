// internal/features/account/error_map_test.go
package account

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	legacyrepo "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

func TestRepositoryErrorsKeepApplicationCodesAndCauses(t *testing.T) {
	failure := errors.New("internal driver details")
	for _, tc := range []struct {
		name string
		err  error
		kind apperr.ErrorKind
	}{
		{"conflict", fmt.Errorf("insert: %w", ErrUserAlreadyExists), apperr.RESOURCE_ALREADY_EXISTS},
		{"not found", fmt.Errorf("update: %w", ErrUserNotFound), apperr.NOT_FOUND},
		{"persistence", errors.Join(repository.ErrRepositoryFailure, failure), apperr.INFRA_DATABASE_ERROR},
		{"legacy persistence", errors.Join(legacyrepo.ErrRepositoryFailure, failure), apperr.INFRA_DATABASE_ERROR},
		{"unknown", failure, apperr.INTERNAL_ERROR},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mapped := mapRepoError("test", tc.err)
			var appErr *apperr.AppError
			if !errors.As(mapped, &appErr) || appErr.Kind != tc.kind || !errors.Is(mapped, tc.err) {
				t.Fatalf("mapped error = %v, want %s with original cause", mapped, tc.kind)
			}
		})
	}

	appErr := apperr.NotFound("existing application error")
	if got := mapRepoError("test", appErr); got != appErr {
		t.Fatal("existing AppError must be preserved")
	}
	if got := mapRepoError("test", nil); got != nil {
		t.Fatalf("nil error became %v", got)
	}
}
