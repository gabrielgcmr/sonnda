// internal/http/middleware/logging.go
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	applog "sonnda-api/internal/app/observability"
)

// AccessLog loga uma linha por request (estilo access log).
func AccessLog(l *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// cria logger contextual por request
		rid, _ := c.Get("request_id")
		route := c.FullPath()
		reqLog := l.With(
			slog.String("request_id", toString(rid)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)
		if route != "" {
			reqLog = reqLog.With(slog.String("route", route))
		}

		// injeta o logger no context da request (usecases/repos podem pegar com logger.FromContext)
		c.Request = c.Request.WithContext(
			applog.IntoContext(c.Request.Context(), reqLog))

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		attrs := []any{
			slog.Int("status", status),
			slog.Int64("latency_ms", latency.Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("response_bytes", c.Writer.Size()),
		}

		// Captura o código de erro de negócio
		if errorCode, exists := c.Get("error_code"); exists {
			attrs = append(attrs, slog.Any("error_code", errorCode))
		}

		if userAgent := c.Request.UserAgent(); userAgent != "" {
			attrs = append(attrs, slog.String("user_agent", userAgent))
		}

		// NOVO: respeita o nível decidido pelo erro (se existir)
		if lvlAny, ok := c.Get("error_log_level"); ok {
			if lvl, ok := lvlAny.(slog.Level); ok {
				switch {
				case status >= 400:
					reqLog.Log(c.Request.Context(), lvl, "request_invalid", attrs...)
				default:
					reqLog.Info("request_completed", attrs...)
				}
				return
			}
		}

		// nível baseado no status
		switch {
		case status >= 500:
			reqLog.Error("request_failed", attrs...)
		case status == http.StatusUnauthorized || status == http.StatusForbidden:
			// 401/403 são comuns; logar como Info evita “poluir” de Warn
			reqLog.Info("request_invalid", attrs...)
		case status >= 400:
			reqLog.Warn("request_invalid", attrs...)
		default:
			reqLog.Info("request_completed", attrs...)
		}
	}
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
