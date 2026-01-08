// internal/app/services/usersvc/error_map.go
package usersvc

import (
	"errors"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/professional"
	"sonnda-api/internal/domain/model/user"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

func mapUserDomainError(err error) error {
	switch {
	case errors.Is(err, user.ErrInvalidAuthProvider),
		errors.Is(err, user.ErrInvalidAuthSubject),
		errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrInvalidFullName),
		errors.Is(err, user.ErrInvalidAccountType),
		errors.Is(err, user.ErrInvalidBirthDate),
		errors.Is(err, user.ErrInvalidCPF),
		errors.Is(err, user.ErrInvalidPhone):
		return &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "dados do usuário inválidos",
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

func mapProfessionalDomainError(err error) error {
	switch {
	case errors.Is(err, professional.ErrRegistrationRequired),
		errors.Is(err, professional.ErrInvalidKind),
		errors.Is(err, professional.ErrInvalidRegistrationNumber),
		errors.Is(err, professional.ErrInvalidRegistrationIssuer):
		return &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "dados do profissional inválidos",
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
