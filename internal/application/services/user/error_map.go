// internal/application/services/user/error_map.go
package usersvc

import (
	"errors"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, user.ErrInvalidAuthIssuer),
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

func mapRepoError(op string, err error) error {
	if err == nil {
		return nil
	}

	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil {
		return appErr
	}

	switch {
	case errors.Is(err, repo.ErrUserAlreadyExists):
		return &apperr.AppError{
			Code:    apperr.RESOURCE_ALREADY_EXISTS,
			Message: "usuário já cadastrado",
			Cause:   err,
		}

	case errors.Is(err, repo.ErrUserNotFound):
		return &apperr.AppError{
			Code:    apperr.NOT_FOUND,
			Message: "usuário não encontrado",
			Cause:   err,
		}

	case errors.Is(err, repo.ErrRepositoryFailure):
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

func userNotFound() error {
	return &apperr.AppError{
		Code:    apperr.NOT_FOUND,
		Message: "usuário não encontrado",
	}
}
