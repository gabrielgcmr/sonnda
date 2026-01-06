// internal/http/errors/error_presenter.go
package errors

import (
	"log/slog"
	"sonnda-api/internal/app/apperr"
	applog "sonnda-api/internal/app/observability"

	"github.com/gin-gonic/gin"
)

func WriteError(c *gin.Context, err error) {
	if c.Writer.Written() {
		c.Abort()
		return
	}

	if err != nil {
		_ = c.Error(err)
	}

	status, resp := ToHTTP(err)
	level := apperr.LogLevelOf(err)

	c.Set("error_code", string(resp.Code))
	c.Set("http_status", status)
	c.Set("error_log_level", level) // slog.Level

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

	// Log quando level >= Warn (429, 413, etc.) OU status >= 500
	if level >= slog.LevelWarn || status >= 500 {
		log.Log(c.Request.Context(), level, "handler_error", attrs...)
	}

	c.Abort()
	c.JSON(status, gin.H{"error": resp})
}
