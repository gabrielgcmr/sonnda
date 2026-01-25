// internal/http/httperr/presenter.go
package apierr

import (
	"errors"
	"fmt"
	"log/slog"
	"sonnda-api/internal/adapters/inbound/http/shared/httperr"
	"sonnda-api/internal/app/apperr"
	applog "sonnda-api/internal/app/observability"

	"github.com/gin-gonic/gin"
)

func ErrorResponder(c *gin.Context, err error) {
	if c.Writer.Written() {
		c.Abort()
		return
	}
	if err != nil {
		_ = c.Error(err)
	}

	status, resp := httperr.ToHTTP(err)
	level := apperr.LogLevelOf(err)

	c.Set("error_code", string(resp.Code))
	c.Set("http_status", status)
	c.Set("error_log_level", level)

	log := applog.FromContext(c.Request.Context())

	attrs := []any{
		slog.Int("status", status),
		slog.String("error_code", string(resp.Code)),
		slog.String("path", c.FullPath()),
		slog.String("method", c.Request.Method),
	}
	if err != nil {
		attrs = append(attrs, slog.Any("err", err))
		if chain := errorChain(err); len(chain) > 0 {
			attrs = append(attrs, slog.Any("error_chain", chain))
		}
	}

	if level >= slog.LevelWarn || status >= 500 {
		log.Log(c.Request.Context(), level, "handler_error", attrs...)
	}

	c.Abort()
	c.JSON(status, gin.H{"error": resp})
}

func errorChain(err error) []string {
	if err == nil {
		return nil
	}

	chain := make([]string, 0, 4)
	seen := map[error]struct{}{}
	for err != nil {
		if _, ok := seen[err]; ok {
			chain = append(chain, "cycle_detected")
			break
		}
		seen[err] = struct{}{}

		chain = append(chain, describeError(err))
		err = errors.Unwrap(err)
	}

	return chain
}

func describeError(err error) string {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil {
		return fmt.Sprintf("app_error code=%s msg=%s", appErr.Code, appErr.Message)
	}
	return err.Error()
}
