// internal/api/handlers/helpers.go
package handlers

import (
	"fmt"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/demographics"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ParseBirthDate valida e converte a data de nascimento do formato ISO (YYYY-MM-DD).
func ParseBirthDate(dateStr string) (time.Time, error) {
	birthDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"birth date must be in YYYY-MM-DD format, got: %s: %w",
			dateStr,
			demographics.ErrInvalidBirthDate,
		)
	}
	return birthDate, nil
}

func parsePatientIDParam(c *gin.Context, id string) (uuid.UUID, bool) {
	idStr := c.Param(id)
	if idStr == "" {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.REQUIRED_FIELD_MISSING,
			Message: "patient_id é obrigatório",
		})
		return uuid.UUID{}, false
	}

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		presenter.ErrorResponder(c, &apperr.AppError{
			Kind:    apperr.INVALID_FIELD_FORMAT,
			Message: "patient_id inválido",
			Cause:   err,
		})
		return uuid.UUID{}, false
	}

	return parsedID, true
}

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
