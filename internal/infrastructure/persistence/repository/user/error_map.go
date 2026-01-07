package user

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"sonnda-api/internal/infrastructure/persistence/repository/helpers"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrRepositoryFailure = errors.New("repository failure")
	ErrNotFound          = errors.New("not found")
)

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	// Não encontrado
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	// Violação de unicidade
	if helpers.IsUniqueViolationError(err) {
		return ErrUserAlreadyExists
	}

	return fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
}
