package handlers

import (
	"fmt"
	"time"

	"sonnda-api/internal/adapters/inbound/http/httperr"
	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/demographics"

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
		httperr.WriteError(c, &apperr.AppError{
			Code:    apperr.REQUIRED_FIELD_MISSING,
			Message: "patient_id é obrigatório",
		})
		return uuid.UUID{}, false
	}

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		httperr.WriteError(c, &apperr.AppError{
			Code:    apperr.INVALID_FIELD_FORMAT,
			Message: "patient_id inválido",
			Cause:   err,
		})
		return uuid.UUID{}, false
	}

	return parsedID, true
}

