package common

import (
	"fmt"
	"sonnda-api/internal/domain/model/shared"
	"time"
)

// ParseBirthDate valida e converte a data de nascimento do formato ISO (YYYY-MM-DD).
func ParseBirthDate(dateStr string) (time.Time, error) {
	birthDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"birth date must be in YYYY-MM-DD format, got: %s: %w",
			dateStr,
			shared.ErrInvalidBirthDate,
		)
	}
	return birthDate, nil
}
