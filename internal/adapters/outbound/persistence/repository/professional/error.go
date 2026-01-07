package professional

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"sonnda-api/internal/adapters/outbound/persistence/repository/helpers"
)

var (
	ErrProfessionalAlreadyExists = errors.New("professional already exists")
	ErrRepositoryFailure         = errors.New("repository failure")
	ErrNotFound                  = errors.New("not found")
)

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if helpers.IsUniqueViolationError(err) {
		return ErrProfessionalAlreadyExists
	}

	return fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
}

