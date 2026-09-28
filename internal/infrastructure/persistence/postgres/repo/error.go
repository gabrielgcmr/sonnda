// internal/infrastructure/persistence/postgres/repo/error.go
// internal/infrastructure/persistence/postgres/repo/error.go
package repo

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

/* ============================================================
   Common errors
   ============================================================ */

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
