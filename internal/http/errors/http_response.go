// internal/app/common/http_response.go
package errors

import (
	"errors"
	"sonnda-api/internal/app/apperr"
)

type ErrorResponse struct {
	Code       apperr.ErrorCode   `json:"code"`
	Message    string             `json:"message"`
	Violations []apperr.Violation `json:"violations,omitempty"`
}

func ToHTTP(err error) (status int, body ErrorResponse) {
	var appErr *apperr.AppError

	if errors.As(err, &appErr) && appErr != nil {
		return StatusFromCode(appErr.Code), ErrorResponse{
			Code:       appErr.Code,
			Message:    appErr.Message,
			Violations: appErr.Violations,
		}
	}

	// Fallback de segurança
	return StatusFromCode(apperr.INTERNAL_ERROR), ErrorResponse{
		Code:    apperr.INTERNAL_ERROR,
		Message: "erro inesperado",
	}
}
