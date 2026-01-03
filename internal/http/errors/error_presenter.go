// internal/http/api/handlers/common/respond_erros.go
package errors

import (
	"log/slog"
	"sonnda-api/internal/app/apperr"
	applog "sonnda-api/internal/app/observability"

	"github.com/gin-gonic/gin"
)

func WriteError(c *gin.Context, err error) {
	if err != nil {
		_ = c.Error(err)
	}

	status, resp := ToHTTP(err)

	c.Set("error_code", string(resp.Code))
	c.Set("http_status", status)
	c.Set("error_log_level", apperr.LogLevelOf(err)) // slog.Level

	log := applog.FromContext(c.Request.Context())

	attrs := []any{
		slog.Int("status", status),
		slog.String("error_code", string(resp.Code)),
		slog.String("path", c.FullPath()),
		slog.String("method", c.Request.Method),
	}

	if err != nil {
		attrs = append(attrs, slog.Any("err", err))
	}

	level := apperr.LogLevelOf(err)
	if status >= 500 {
		log.Log(c.Request.Context(), level, "handler_error", attrs...)
	}
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    resp.Code,
			"message": resp.Message,
		},
	})

}
