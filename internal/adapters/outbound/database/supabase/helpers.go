package supabase

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

/* ============================================================
   NULL HELPERS
   ============================================================ */

func nullableString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func nullableTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time
	return &t
}
