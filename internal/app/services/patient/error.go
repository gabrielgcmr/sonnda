package patientsvc

import (
	"errors"
	"fmt"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/demographics"
	"sonnda-api/internal/domain/model/patient"
)

var (
	ErrAuthorizationForbidden = errors.New("authorization forbidden")
	ErrPatientNotFound        = errors.New("patient not found")
	ErrCPFAlreadyExists       = errors.New("cpf already exists")
)

func mapPatientDomainError(err error) error {
	switch {
	// authorization
	case errors.Is(err, ErrAuthorizationForbidden):
		return &apperr.AppError{
			Code:    apperr.ACCESS_DENIED,
			Message: "acesso negado",
			Cause:   err,
		}

	// validation
	case errors.Is(err, patient.ErrInvalidFullName),
		errors.Is(err, demographics.ErrInvalidBirthDate),
		errors.Is(err, demographics.ErrInvalidCPF),
		errors.Is(err, demographics.ErrInvalidGender),
		errors.Is(err, demographics.ErrInvalidRace),
		errors.Is(err, patient.ErrInvalidBirthDate),
		errors.Is(err, patient.ErrInvalidGender),
		errors.Is(err, patient.ErrInvalidRace):
		return &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "dados inv\u00e1lidos",
			Cause:   err,
		}

	// conflict
	case errors.Is(err, ErrCPFAlreadyExists):
		return &apperr.AppError{
			Code:    apperr.RESOURCE_ALREADY_EXISTS,
			Message: "paciente j\u00e1 cadastrado",
			Cause:   err,
		}

	// not found
	case errors.Is(err, ErrPatientNotFound):
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
