// internal/application/services/patient/error.go
package patientsvc

import (
	"errors"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/demographics"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
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
		return apperr.Validation("dados inválidos")

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
	case errors.Is(err, repo.ErrPatientAlreadyExists):
		return apperr.AlreadyExists("paciente já cadastrado")

	case errors.Is(err, repo.ErrPatientNotFound):
		return patientNotFound()

	case errors.Is(err, repo.ErrRepositoryFailure):
		return apperr.Internal("falha técnica", fmt.Errorf("%s: %w", op, err))

	default:
		return apperr.Internal("erro inesperado", fmt.Errorf("%s: %w", op, err))
	}
}

func patientNotFound() error {
	return apperr.NotFound("paciente não encontrado")
}
