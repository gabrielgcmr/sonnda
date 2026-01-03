// internal/app/common/http_response.go
package errors

import (
	"errors"
	"sonnda-api/internal/app/apperr"
)

type ErrorResponse struct {
	Code    apperr.ErrorCode `json:"code"`
	Message string           `json:"message"`
}

func ToHTTP(err error) (status int, body ErrorResponse) {
	var appErr *apperr.AppError

	if errors.As(err, &appErr) && appErr != nil {
		return HTTPStatus(appErr.Code), ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
		}
	}

	// Fallback de segurança
	return HTTPStatus(apperr.INTERNAL_ERROR), ErrorResponse{
		Code:    apperr.INTERNAL_ERROR,
		Message: "erro inesperado",
	}
}
