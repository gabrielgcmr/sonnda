// internal/app/common/http_mapper.go
package httperr

import (
	"errors"
	"net/http"
	"github.com/gabrielgcmr/sonnda/internal/app/apperr"
)

func StatusFromCode(code apperr.ErrorCode) int {
	switch code {

	// AUTH
	case apperr.AUTH_REQUIRED,
		apperr.AUTH_TOKEN_INVALID,
		apperr.AUTH_TOKEN_EXPIRED:
		return http.StatusUnauthorized // 401

	// AUTHZ
	case apperr.ACCESS_DENIED,
		apperr.ACTION_NOT_ALLOWED:
		return http.StatusForbidden // 403

	// VALIDATION
	case apperr.VALIDATION_FAILED:
		return http.StatusBadRequest // 400

	// NOT FOUND
	case apperr.NOT_FOUND:
		return http.StatusNotFound // 404

	// CONFLICT
	case apperr.RESOURCE_CONFLICT,
		apperr.RESOURCE_ALREADY_EXISTS:
		return http.StatusConflict // 409

	// DOMAIN
	case apperr.DOMAIN_RULE_VIOLATION:
		return http.StatusUnprocessableEntity // 422

	// RATE
	case apperr.RATE_LIMIT_EXCEEDED:
		return http.StatusTooManyRequests // 429
	case apperr.UPLOAD_SIZE_EXCEEDED:
		return http.StatusRequestEntityTooLarge // 413

	// INFRA
	case apperr.INFRA_EXTERNAL_SERVICE_ERROR:
		return http.StatusBadGateway // 502
	case apperr.INFRA_TIMEOUT:
		return http.StatusGatewayTimeout // 504
	case apperr.INFRA_AUTHENTICATION_ERROR,
		apperr.INFRA_DATABASE_ERROR,
		apperr.INFRA_STORAGE_ERROR:
		return http.StatusInternalServerError // 500

	// INTERNAL
	case apperr.INTERNAL_ERROR:
		fallthrough
	default:
		return http.StatusInternalServerError // 500
	}
}

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
