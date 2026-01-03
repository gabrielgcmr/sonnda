// internal/app/common/http_mapper.go
package errors

import (
	"net/http"
	"sonnda-api/internal/app/apperr"
)

func HTTPStatus(code apperr.ErrorCode) int {
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
	case apperr.VALIDATION_FAILED,
		apperr.REQUIRED_FIELD_MISSING,
		apperr.INVALID_FIELD_FORMAT,
		apperr.INVALID_ENUM_VALUE,
		apperr.INVALID_DATE:
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
