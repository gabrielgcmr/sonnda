package supabase

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
