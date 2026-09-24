// internal/infrastructure/persistence/postgres/repo/error.go
package repo

import (
	"errors"

	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

/* ============================================================
   Common errors
   ============================================================ */

var (
	//Common
	ErrRepositoryFailure = repository.ErrRepositoryFailure
	//Professional
	ErrProfessionalAlreadyExists = errors.New("professional already exists")
	ErrProfessionalNotFound      = errors.New("professional not found")
	//patient
	ErrPatientAlreadyExists = errors.New("patient already exists")
	ErrPatientNotFound      = errors.New("patient not found")
)

func IsUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func IsPgNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
