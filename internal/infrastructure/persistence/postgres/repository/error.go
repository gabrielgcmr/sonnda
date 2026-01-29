// internal/adapters/outbound/persistence/postgres/repository/error.go
package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

/* ============================================================
   Common errors
   ============================================================ */

var (
	//Common
	ErrRepositoryFailure = errors.New("repository failure")
	//User
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
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
