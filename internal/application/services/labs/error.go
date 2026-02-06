// internal/application/services/labs/error.go
package labsvc

import (
	"errors"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

func mapRepoError(op string, err error) error {
	if err == nil {
		return nil
	}

	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil {
		return appErr
	}

	// A maioria dos erros aqui vem de DB, então tratamos como infra por padrão.
	// Se quisermos diferenciar depois, dá pra adicionar casos específicos.
	switch {
	case errors.Is(err, repo.ErrRepositoryFailure):
		fallthrough
	default:
		return &apperr.AppError{
			Kind:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   fmt.Errorf("%s: %w", op, err),
		}
	}
}

func patientNotFound() error {
	return &apperr.AppError{
		Kind:    apperr.NOT_FOUND,
		Message: "paciente não encontrado",
	}
}
