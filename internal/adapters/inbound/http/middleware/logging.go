package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	obslog "sonnda-api/internal/observability/logger"
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
func AccessLog(base *slog.Logger, getUserID func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// cria logger contextual por request
		rid, _ := c.Get("request_id")
		reqLog := base.With(
			slog.String("request_id", toString(rid)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
		)

		// coloca no context pra usecases/repos pegarem
		c.Request = c.Request.WithContext(obslog.IntoContext(c.Request.Context(), reqLog))

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		attrs := []slog.Attr{
			slog.Int("status", status),
			slog.Int64("latency_ms", latency.Milliseconds()),
		}
		if getUserID != nil {
			if uid := getUserID(c); uid != "" {
				attrs = append(attrs, slog.String("user_id", uid))
			}
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("gin_errors", c.Errors.String()))
		}

		// nível baseado no status
		switch {
		case status >= 500:
			reqLog.Error("request_completed", attrsToAny(attrs)...)
		case status >= 400:
			reqLog.Warn("request_completed", attrsToAny(attrs)...)
		default:
			reqLog.Info("request_completed", attrsToAny(attrs)...)
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

func attrsToAny(attrs []slog.Attr) []any {
	out := make([]any, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, a)
	}
	return out
}
