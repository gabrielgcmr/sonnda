package patient

import (
	"fmt"
	"time"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/entities/shared"
)

// ParseBirthDate valida e converte a data de nascimento do formato ISO (YYYY-MM-DD).
func ParseBirthDate(dateStr string) (time.Time, error) {
	birthDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, apperr.New(
			apperr.KindInvalidInput,
			"invalid_birth_date",
			fmt.Sprintf("birth date must be in YYYY-MM-DD format, got: %s", dateStr),
		)
	}
	return birthDate, nil
}

// ParseGender valida e converte o gênero para o tipo do domínio.
func ParseGender(genderStr string) (shared.Gender, error) {
	gender, err := shared.ParseGender(genderStr)
	if err != nil {
		return "", apperr.New(
			apperr.KindInvalidInput,
			"invalid_gender",
			fmt.Sprintf("invalid gender value: %s", genderStr),
		)
	}
	return gender, nil
}

// ParseRace valida e converte a raça/etnia para o tipo do domínio.
func ParseRace(raceStr string) (shared.Race, error) {
	race, err := shared.ParseRace(raceStr)
	if err != nil {
		return "", apperr.New(
			apperr.KindInvalidInput,
			"invalid_race",
			fmt.Sprintf("invalid race value: %s", raceStr),
		)
	}
	return race, nil
}
