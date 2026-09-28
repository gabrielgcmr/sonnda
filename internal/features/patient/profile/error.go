// internal/features/patient/profile/error.go
package patientprofile

import (
	"errors"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/domain/demographics"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/persistence"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, profiledomain.ErrInvalidFullName),
		errors.Is(err, demographics.ErrInvalidBirthDate),
		errors.Is(err, demographics.ErrInvalidCPF),
		errors.Is(err, demographics.ErrInvalidGender),
		errors.Is(err, demographics.ErrInvalidRace),
		errors.Is(err, profiledomain.ErrInvalidBirthDate),
		errors.Is(err, profiledomain.ErrInvalidGender),
		errors.Is(err, profiledomain.ErrInvalidRace):
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
	case errors.Is(err, ErrPatientAlreadyExists):
		return apperr.AlreadyExists("paciente já cadastrado")

	case errors.Is(err, ErrPatientNotFound):
		return patientNotFound()

	case errors.Is(err, persistence.ErrPersistenceFailure):
		return apperr.Internal("falha técnica", fmt.Errorf("%s: %w", op, err))

	default:
		return apperr.Internal("erro inesperado", fmt.Errorf("%s: %w", op, err))
	}
}

func patientNotFound() error {
	return apperr.NotFound("paciente não encontrado")
}
