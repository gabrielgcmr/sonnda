package binder

import (
	"github.com/gabrielgcmr/sonnda/internal/app/apperr"

	"github.com/gin-gonic/gin"
)

func BindJSON(c *gin.Context, dst any) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		violations := ValidationErrorsToViolations(err)

		appErr := &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "dados de entrada inválidos",
			Cause:   err,
		}

		if len(violations) > 0 {
			appErr.Violations = violations
		}

		return appErr
	}
	return nil
}
