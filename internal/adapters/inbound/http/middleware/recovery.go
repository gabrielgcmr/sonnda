package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	applog "sonnda-api/internal/app/observability"
)

// Recovery captura panics, loga stacktrace estruturado e devolve 500 sem derrubar o servidor.
func Recovery(l *slog.Logger) gin.HandlerFunc {
	if l == nil {
		l = slog.Default()
	}

	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				c.Set("error_code", "panic")

				rid, _ := c.Get("request_id")
				route := c.FullPath()

				attrs := []any{
					slog.String("request_id", toString(rid)),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("client_ip", c.ClientIP()),
					slog.String("user_agent", c.Request.UserAgent()),
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				}
				if route != "" {
					attrs = append(attrs, slog.String("route", route))
				}
				if u, ok := GetCurrentUser(c); ok && u != nil {
					attrs = append(attrs, slog.String("user_id", u.ID.String()))
				}

				// Se o AccessLog já injetou um logger no context, usamos ele; senão usamos o logger recebido.
				reqLog := applog.FromContext(c.Request.Context())
				if reqLog == nil {
					reqLog = l
				}
				reqLog.Error("panic_recovered", attrs...)

				if !c.Writer.Written() {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
					return
				}
				c.Abort()
			}
		}()

		c.Next()
	}
}
