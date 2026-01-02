// internal/http/api/handlers/common/error_mapper.go
package common

import (
	"net/http"

	"sonnda-api/internal/app/apperr"

	"github.com/gin-gonic/gin"
)

func RespondAppError(c *gin.Context, err error) {
	ae, ok := apperr.As(err)
	if !ok {
		RespondDomainError(c, err)
		return
	}

	switch ae.Kind {

	case apperr.KindInvalidInput:
		RespondError(c, http.StatusBadRequest, ae.Code, err)

	case apperr.KindNotFound:
		RespondError(c, http.StatusNotFound, ae.Code, nil)

	case apperr.KindConflict:
		RespondError(c, http.StatusConflict, ae.Code, nil)

	case apperr.KindUnauthorized:
		RespondError(c, http.StatusUnauthorized, ae.Code, nil)

	case apperr.KindForbidden:
		RespondError(c, http.StatusForbidden, ae.Code, nil)

	case apperr.KindUnavailable, apperr.KindServiceClosed:
		RespondError(c, http.StatusServiceUnavailable, ae.Code, nil)

	case apperr.KindTimeout:
		RespondError(c, http.StatusGatewayTimeout, ae.Code, err)

	case apperr.KindBadGateway, apperr.KindExternal:
		RespondError(c, http.StatusBadGateway, ae.Code, err)

	default:
		RespondError(c, http.StatusInternalServerError, "server_error", err)
	}
}
