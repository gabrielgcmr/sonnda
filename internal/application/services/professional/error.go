// internal/application/services/professional/error.go
package professionalsvc

import (
	"errors"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/professional"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, professional.ErrInvalidUserID),
		errors.Is(err, professional.ErrInvalidKind),
		errors.Is(err, professional.ErrInvalidRegistrationNumber),
		errors.Is(err, professional.ErrInvalidRegistrationIssuer):
		return apperr.Validation("dados profissionais inválidos")

	default:
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr != nil {
			return appErr
		}
		return apperr.Internal("erro inesperado", err)
	}
}

func mapRepoError(op string, err error) error {
	if err == nil {
		return nil
	}

	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil {
		return appErr
	}

	switch {
	case errors.Is(err, repo.ErrProfessionalAlreadyExists):
		return apperr.AlreadyExists("profissional já cadastrado")

	case errors.Is(err, repo.ErrProfessionalNotFound):
		return professionalNotFound()

	case errors.Is(err, repo.ErrRepositoryFailure):
		return &apperr.AppError{
			Kind:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   fmt.Errorf("%s: %w", op, err),
		}

	default:
		return apperr.Internal("erro inesperado", fmt.Errorf("%s: %w", op, err))
	}
}

func professionalNotFound() error {
	return apperr.NotFound("profissional não encontrado")
}
