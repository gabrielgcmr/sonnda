package patientsvc

import (
	"errors"
	"fmt"

	repoerr "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/demographics"
	"sonnda-api/internal/domain/model/patient"
)

func mapDomainError(err error) error {
	switch {
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
			Message: "dados inválidos",
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
	case errors.Is(err, repoerr.ErrPatientAlreadyExists):
		return &apperr.AppError{
			Code:    apperr.RESOURCE_ALREADY_EXISTS,
			Message: "paciente já cadastrado",
			Cause:   err,
		}

	case errors.Is(err, repoerr.ErrPatientNotFound):
		return &apperr.AppError{
			Code:    apperr.NOT_FOUND,
			Message: "paciente não encontrado",
			Cause:   err,
		}

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

func patientNotFound() error {
	return &apperr.AppError{
		Code:    apperr.NOT_FOUND,
		Message: "paciente não encontrado",
	}
}
