package patientsvc

import (
	"errors"
	"fmt"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/shared"
	"sonnda-api/internal/domain/model/user"
)

func mapPatientDomainError(err error) error {
	switch {
	// authorization
	case errors.Is(err, user.ErrAuthorizationForbidden):
		return &apperr.AppError{
			Code:    apperr.ACCESS_DENIED,
			Message: "acesso negado",
			Cause:   err,
		}

	// validation
	case errors.Is(err, patient.ErrInvalidFullName),
		errors.Is(err, shared.ErrInvalidBirthDate),
		errors.Is(err, shared.ErrInvalidCPF),
		errors.Is(err, shared.ErrInvalidGender),
		errors.Is(err, shared.ErrInvalidRace),
		errors.Is(err, patient.ErrInvalidBirthDate),
		errors.Is(err, patient.ErrInvalidGender),
		errors.Is(err, patient.ErrInvalidRace):
		return &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "dados inv\u00e1lidos",
			Cause:   err,
		}

	// conflict
	case errors.Is(err, patient.ErrCPFAlreadyExists):
		return &apperr.AppError{
			Code:    apperr.RESOURCE_ALREADY_EXISTS,
			Message: "paciente j\u00e1 cadastrado",
			Cause:   err,
		}

	// not found
	case errors.Is(err, patient.ErrPatientNotFound):
		return &apperr.AppError{
			Code:    apperr.NOT_FOUND,
			Message: "paciente n\u00e3o encontrado",
			Cause:   err,
		}

	default:
		return &apperr.AppError{
			Code:    apperr.INTERNAL_ERROR,
			Message: "erro inesperado",
			Cause:   err,
		}
	}
}

func mapInfraError(op string, err error) error {
	return &apperr.AppError{
		Code:    apperr.INFRA_DATABASE_ERROR,
		Message: "falha t\u00e9cnica",
		Cause:   fmt.Errorf("%s: %w", op, err),
	}
}
