// internal/app/services/usersvc/error_map.go
package usersvc

import (
	"errors"
	"fmt"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/model/user/professional"
)

func mapUserDomainError(err error) error {
	switch {
	case errors.Is(err, user.ErrInvalidAuthProvider),
		errors.Is(err, user.ErrInvalidAuthSubject),
		errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrInvalidFullName),
		errors.Is(err, user.ErrInvalidBirthDate),
		errors.Is(err, user.ErrInvalidCPF),
		errors.Is(err, user.ErrInvalidPhone):
		return &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "dados inválidos",
			Cause:   err,
		}

	case errors.Is(err, user.ErrEmailAlreadyExists),
		errors.Is(err, user.ErrCPFAlreadyExists),
		errors.Is(err, user.ErrAuthIdentityAlreadyExists):
		return &apperr.AppError{
			Code:    apperr.RESOURCE_ALREADY_EXISTS,
			Message: "usuário já cadastrado",
			Cause:   err,
		}

	case errors.Is(err, user.ErrUserNotFound):
		return &apperr.AppError{
			Code:    apperr.NOT_FOUND,
			Message: "usuário não encontrado",
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
		errors.Is(err, professional.ErrInvalidRegistrationNumber),
		errors.Is(err, professional.ErrInvalidRegistrationIssuer):
		return &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "dados profissionais inválidos",
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

func mapInfraError(msg string, err error) error {
	// msg aqui é para contexto de log interno, não precisa vazar no HTTP.
	return &apperr.AppError{
		Code:    apperr.INFRA_DATABASE_ERROR,
		Message: "falha técnica",
		Cause:   fmt.Errorf("%s: %w", msg, err),
	}
}
