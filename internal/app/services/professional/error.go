package professionalsvc

import (
	"errors"
	"fmt"

	repoerr "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/professional"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, professional.ErrInvalidUserID),
		errors.Is(err, professional.ErrInvalidKind),
		errors.Is(err, professional.ErrInvalidRegistrationNumber),
		errors.Is(err, professional.ErrInvalidRegistrationIssuer):
		return &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "dados profissionais inválidos",
			Cause:   err,
		}

	default:
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr != nil {
			return appErr
		}
		return &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "erro inesperado",
			Cause:   err,
		}
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
	case errors.Is(err, repoerr.ErrProfessionalAlreadyExists):
		return &apperr.AppError{
			Code:    apperr.RESOURCE_ALREADY_EXISTS,
			Message: "profissional já cadastrado",
			Cause:   err,
		}

	case errors.Is(err, repoerr.ErrProfessionalNotFound):
		return professionalNotFound(err)

	case errors.Is(err, repoerr.ErrRepositoryFailure):
		return &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   fmt.Errorf("%s: %w", op, err),
		}

	default:
		return &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "erro inesperado",
			Cause:   fmt.Errorf("%s: %w", op, err),
		}
	}
}

func professionalNotFound(cause error) error {
	return &apperr.AppError{
		Code:    apperr.NOT_FOUND,
		Message: "profissional não encontrado",
		Cause:   cause,
	}
}
