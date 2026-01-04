package errors

import (
	"sonnda-api/internal/app/apperr"

	"github.com/gin-gonic/gin"
)

func BindJSON(c *gin.Context, dst any) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		violations := ValidationErrorsToViolations(err)

		appErr := &apperr.AppError{
			Code:    apperr.VALIDATION_FAILED,
			Message: "validação falhou",
			Cause:   err,
		}

		if len(violations) > 0 {
			appErr.Violations = violations
		}

		return appErr
	}
	return nil
}
