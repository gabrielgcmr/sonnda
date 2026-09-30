// internal/api/humaerror/huma_error.go
package humaerror

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

// From translates an application error into Huma's RFC 9457 error model.
func From(err error) *huma.ErrorModel {
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) || appErr == nil {
		return huma.NewError(http.StatusInternalServerError, "erro inesperado").(*huma.ErrorModel)
	}

	details := make([]error, 0, len(appErr.Violations))
	for _, violation := range appErr.Violations {
		details = append(details, &huma.ErrorDetail{
			Location: violation.Field,
			Message:  violation.Reason,
		})
	}

	return huma.NewError(StatusFromKind(appErr.Kind), appErr.Message, details...).(*huma.ErrorModel)
}

// Write writes an application error using Huma's configured error writer.
func Write(api huma.API, ctx huma.Context, err error) error {
	model := From(err)
	details := make([]error, len(model.Errors))
	for i, detail := range model.Errors {
		details[i] = detail
	}
	return huma.WriteErr(api, ctx, model.Status, model.Detail, details...)
}

// StatusFromKind maps the stable application error code to an HTTP status.
func StatusFromKind(kind apperr.ErrorKind) int {
	switch kind {
	case apperr.AUTH_REQUIRED, apperr.AUTH_TOKEN_INVALID, apperr.AUTH_TOKEN_EXPIRED:
		return http.StatusUnauthorized
	case apperr.PROFILE_NOT_FOUND, apperr.ACCESS_DENIED, apperr.ACTION_NOT_ALLOWED:
		return http.StatusForbidden
	case apperr.VALIDATION_FAILED, apperr.REQUIRED_FIELD_MISSING, apperr.INVALID_FIELD_FORMAT, apperr.INVALID_ENUM_VALUE, apperr.INVALID_DATE:
		return http.StatusBadRequest
	case apperr.NOT_FOUND:
		return http.StatusNotFound
	case apperr.RESOURCE_CONFLICT, apperr.RESOURCE_ALREADY_EXISTS:
		return http.StatusConflict
	case apperr.DOMAIN_RULE_VIOLATION:
		return http.StatusUnprocessableEntity
	case apperr.RATE_LIMIT_EXCEEDED:
		return http.StatusTooManyRequests
	case apperr.UPLOAD_SIZE_EXCEEDED:
		return http.StatusRequestEntityTooLarge
	case apperr.INFRA_EXTERNAL_SERVICE_ERROR:
		return http.StatusBadGateway
	case apperr.INFRA_TIMEOUT:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
