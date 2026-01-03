package patient

import (
	"fmt"
	"time"

	"sonnda-api/internal/domain/model/shared"
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

// ParseGender valida e converte o gênero para o tipo do domínio.
func ParseGender(genderStr string) (shared.Gender, error) {
	gender, err := shared.ParseGender(genderStr)
	if err != nil {
		return "", fmt.Errorf("invalid gender value: %s: %w", genderStr, shared.ErrInvalidGender)
	}
	return gender, nil
}

// ParseRace valida e converte a raça/etnia para o tipo do domínio.
func ParseRace(raceStr string) (shared.Race, error) {
	race, err := shared.ParseRace(raceStr)
	if err != nil {
		return "", fmt.Errorf("invalid race value: %s: %w", raceStr, shared.ErrInvalidRace)
	}
	return race, nil
}
