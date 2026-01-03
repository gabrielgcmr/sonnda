package patient

import (
	"fmt"

	"sonnda-api/internal/domain/model/demographics"
)

// ParseGender valida e converte o gênero para o tipo do domínio.
func ParseGender(genderStr string) (demographics.Gender, error) {
	gender, err := demographics.ParseGender(genderStr)
	if err != nil {
		return "", fmt.Errorf("invalid gender value: %s: %w", genderStr, demographics.ErrInvalidGender)
	}
	return gender, nil
}

// ParseRace valida e converte a raça/etnia para o tipo do domínio.
func ParseRace(raceStr string) (demographics.Race, error) {
	race, err := demographics.ParseRace(raceStr)
	if err != nil {
		return "", fmt.Errorf("invalid race value: %s: %w", raceStr, demographics.ErrInvalidRace)
	}
	return race, nil
}
