// internal/features/patient/exam/laboratory/postgres/errors.go
package postgres

import (
	"errors"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/kernel/persistence"
	"github.com/jackc/pgx/v5"
)

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func persistenceError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(persistence.ErrPersistenceFailure, fmt.Errorf("%s: %w", operation, err))
}
