package helpers

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

func IsPgNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

/* ============================================================
   Text conversions (*string <-> pgtype.Text)
   ============================================================ */

func FromNullableStringToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// FromRequiredStringToPgText converts non-null string to pgtype.Text.
func FromRequiredStringToPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func FromPgTextToNullableString(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}
func FromPgTextToRequiredString(t pgtype.Text) (string, error) {
	if !t.Valid {
		return "", fmt.Errorf("pgtype.Text is NULL/invalid")
	}
	return t.String, nil
}

/* ============================================================
   Date conversions (*time.Time <-> pgtype.Date)
   ============================================================ */

func FromNullableDateToPgDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func FromRequiredDateToPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func FromPgDateToNullableDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	tt := d.Time
	return &tt
}
func FromPgDateToRequiredDate(d pgtype.Date) (time.Time, error) {
	if !d.Valid {
		return time.Time{}, fmt.Errorf("pgtype.Date is NULL/invalid")
	}
	return d.Time, nil
}

/* ============================================================
   Timestamptz conversions (*time.Time <-> pgtype.Timestamptz)
   ============================================================ */

func FromNullableTimestamptzToPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func FromRequiredTimestamptzToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func FromPgTimestamptzToNullableTimestamptz(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	tt := ts.Time
	return &tt
}
func FromPgTimestamptzToRequiredTimestamptz(ts pgtype.Timestamptz) (time.Time, error) {
	if !ts.Valid {
		return time.Time{}, fmt.Errorf("pgtype.Timestamptz is NULL/invalid")
	}
	return ts.Time, nil
}

/* ============================================================
   UUID conversions (*uuid.UUID <-> pgtype.Text)
   ============================================================ */

func FromNullableUUIDToPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}

	var bytes [16]byte
	copy(bytes[:], id[:])

	return pgtype.UUID{
		Bytes: bytes,
		Valid: true,
	}
}

func FromPgUUIDToNullableUUID(u pgtype.UUID) (*uuid.UUID, error) {
	if !u.Valid {
		return nil, nil
	}
	parsed, err := uuid.FromBytes(u.Bytes[:])
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func FromNullableUUIDToPgText(id *uuid.UUID) pgtype.Text {
	if id == nil || *id == uuid.Nil {
		return pgtype.Text{Valid: false}
	}
	value := id.String()
	return FromNullableStringToPgText(&value)
}

func FromPgTextToNullableUUID(t pgtype.Text) (*uuid.UUID, error) {
	if !t.Valid {
		return nil, nil
	}
	raw := strings.TrimSpace(t.String)
	if raw == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
