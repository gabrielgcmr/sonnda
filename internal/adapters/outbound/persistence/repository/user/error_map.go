package user

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
)

var (
	ErrAlreadyExists     = errors.New("user already exists")
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
	var pgErr *pgconn.PgError
	// Violação de unicidade
	if errors.As(err, &pgErr) {
		return ErrAlreadyExists
	}

	return fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
}
