package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	applog "sonnda-api/internal/logger"
)

const requestIDHeader = "X-Request-ID"

// RequestID garante que cada request tenha um request_id e propaga em header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(requestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Writer.Header().Set(requestIDHeader, rid)
		c.Set("request_id", rid)
		c.Next()
	}
}

// AccessLog loga uma linha por request (estilo access log).
func AccessLog(l *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// cria logger contextual por request
		rid, _ := c.Get("request_id")
		reqLog := l.With(
			slog.String("request_id", toString(rid)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
		)

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
		}

		if userAgent := c.Request.UserAgent(); userAgent != "" {
			attrs = append(attrs, slog.String("user_agent", userAgent))
		}

		if u, ok := CurrentUser(c); ok && u != nil {
			attrs = append(attrs, slog.String("user_id", u.ID.String()))
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("gin_errors", c.Errors.String()))
		}

		// nível baseado no status
		switch {
		case status >= 500:
			reqLog.Error("request_completed", attrs...)
		case status == http.StatusUnauthorized || status == http.StatusForbidden:
			// 401/403 são comuns; logar como Info evita “poluir” de Warn
			reqLog.Info("request_completed", attrs...)
		case status >= 400:
			reqLog.Warn("request_completed", attrs...)
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
