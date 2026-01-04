package professionalsvc

import (
	"errors"
	"fmt"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/user/professional"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, professional.ErrInvalidUserID),
		errors.Is(err, professional.ErrInvalidKind),
		errors.Is(err, professional.ErrInvalidRegistrationNumber),
		errors.Is(err, professional.ErrInvalidRegistrationIssuer),
		errors.Is(err, professional.ErrRegistrationRequired):
		return &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "dados profissionais inválidos",
			Cause:   err,
		}

	case errors.Is(err, professional.ErrProfileNotFound):
		return &apperr.AppError{
			Code:    apperr.NOT_FOUND,
			Message: "profissional não encontrado",
			Cause:   err,
		}

	default:
		var appErr *apperr.AppError
		if errors.As(err, &appErr) {
			return err
		}
		return &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "erro inesperado",
			Cause:   err,
		}
	}
}

func mapInfraError(op string, err error) error {
	if err == nil {
		return nil
	}
	return &apperr.AppError{
		Code:    apperr.INFRA_DATABASE_ERROR,
		Message: "falha técnica",
		Cause:   fmt.Errorf("%s: %w", op, err),
	}
}
