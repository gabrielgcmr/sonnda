package patient

import (
	"fmt"

	"sonnda-api/internal/domain/model/shared"
)

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
