package supabase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

/* ============================================================
   Common errors
   ============================================================ */

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

/* ============================================================
   UUID parsing (string -> uuid.UUID)
   ============================================================ */

// ParseUUID parses a UUID from string (trimmed). Returns error if invalid.
func ParseUUID(s string) (uuid.UUID, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return uuid.Nil, fmt.Errorf("uuid is empty")
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid %q: %w", s, err)
	}
	return u, nil
}

// ParseUUIDPtr parses UUID from *string. Nil/empty => nil.
func ParseUUIDPtr(s *string) (*uuid.UUID, error) {
	if s == nil {
		return nil, nil
	}
	str := strings.TrimSpace(*s)
	if str == "" {
		return nil, nil
	}
	u, err := uuid.Parse(str)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %q: %w", str, err)
	}
	return &u, nil
}

/* ============================================================
   UUID conversions (uuid.UUID <-> pgtype.UUID)
   ============================================================ */

// ToPgUUID converts non-null uuid.UUID to pgtype.UUID.
func ToPgUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

// ToPgUUIDPtr converts optional *uuid.UUID to pgtype.UUID (nullable).
func ToPgUUIDPtr(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

// ToPgUUIDFromStringPtr converts optional *string (uuid text) to pgtype.UUID (nullable).
func ToPgUUIDFromStringPtr(s *string) (pgtype.UUID, error) {
	u, err := ParseUUIDPtr(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return ToPgUUIDPtr(u), nil
}

// FromPgUUID converts pgtype.UUID to uuid.UUID.
// Use only when DB column is NOT NULL (Valid should be true).
func FromPgUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return uuid.UUID(u.Bytes)
}

// FromPgUUIDPtr converts pgtype.UUID (nullable) to *uuid.UUID.
func FromPgUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	v := uuid.UUID(u.Bytes)
	return &v
}

// FromPgUUIDToStringPtr converts pgtype.UUID (nullable) to *string (uuid text).
func FromPgUUIDToStringPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

/* ============================================================
   Text conversions (*string <-> pgtype.Text)
   ============================================================ */

func ToText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func FromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

// ToTextValue converts non-null string to pgtype.Text.
func ToTextValue(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// ToTextValuePtr converts optional *string to pgtype.Text (nullable).
// (você já tem ToText, mas esse nome deixa mais claro)
func ToTextValuePtr(s *string) pgtype.Text {
	return ToText(s)
}

/* ============================================================
   Date conversions (*time.Time <-> pgtype.Date)
   ============================================================ */

func ToDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func FromDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	tt := d.Time
	return &tt
}

/* ============================================================
   Timestamptz conversions (*time.Time <-> pgtype.Timestamptz)
   ============================================================ */

func ToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func FromTimestamptz(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	tt := ts.Time
	return &tt
}

// MustTime extracts time.Time from pgtype.Timestamptz when you expect NOT NULL.
func MustTime(ts pgtype.Timestamptz) (time.Time, error) {
	if !ts.Valid {
		return time.Time{}, fmt.Errorf("timestamptz is NULL/invalid")
	}
	return ts.Time, nil
}
