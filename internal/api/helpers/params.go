// internal/api/helpers/params.go
package helpers

import (
	"strconv"

	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ParsePatientIDParam(c *gin.Context, parameterName string) (uuid.UUID, bool) {
	idStr := c.Param(parameterName)
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

func ParsePagination(c *gin.Context, defaultLimit, defaultOffset int) (limit, offset int, ok bool) {
	limit = defaultLimit
	offset = defaultOffset

	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed <= 0 {
			presenter.ErrorResponder(c, &apperr.AppError{
				Kind:    apperr.VALIDATION_FAILED,
				Message: "limit deve ser > 0",
				Cause:   err,
			})
			return 0, 0, false
		}
		limit = parsed
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			presenter.ErrorResponder(c, &apperr.AppError{
				Kind:    apperr.VALIDATION_FAILED,
				Message: "offset deve ser >= 0",
				Cause:   err,
			})
			return 0, 0, false
		}
		offset = parsed
	}

	return limit, offset, true
}
