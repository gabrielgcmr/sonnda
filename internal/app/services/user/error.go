// internal/app/services/usersvc/error_map.go
package usersvc

import (
	"errors"
	"fmt"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/model/user/professional"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrCPFAlreadyExists   = errors.New("cpf already exists")
	ErrUserNotFound       = errors.New("user not found")
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
		return apperr.Validation("dados inválidos", apperr.Violation{Reason: err.Error()})

	case errors.Is(err, ErrEmailAlreadyExists),
		errors.Is(err, ErrCPFAlreadyExists),
		errors.Is(err, user.ErrAuthIdentityAlreadyExists):
		return apperr.Conflict("usuário já cadastrado")

	case errors.Is(err, ErrUserNotFound):
		return &apperr.AppError{
			Code:    apperr.NOT_FOUND,
			Message: "usuário não encontrado",
			Cause:   err,
		}

	default:
		return apperr.Internal("erro inesperado", err)
	}
}

func mapProfessionalDomainError(err error) error {
	switch {
	case errors.Is(err, professional.ErrRegistrationRequired),
		errors.Is(err, professional.ErrInvalidKind),
		errors.Is(err, professional.ErrInvalidRegistrationNumber),
		errors.Is(err, professional.ErrInvalidRegistrationIssuer):
		return apperr.Validation("dados profissionais inválidos", apperr.Violation{Reason: err.Error()})
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
