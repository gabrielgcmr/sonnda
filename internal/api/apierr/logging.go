// internal/api/apierr/logging.go
package apierr

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/gin-gonic/gin"
)

func buildLogAttrs(c *gin.Context, status int, code apperr.ErrorCode, err error) []any {
	attrs := []any{
		slog.Int("status", status),
		slog.String("error_code", string(code)),
		slog.String("path", c.FullPath()),
		slog.String("method", c.Request.Method),
	}

	if err != nil {
		attrs = append(attrs, slog.Any("err", err))
		if chain := errorChain(err); len(chain) > 0 {
			attrs = append(attrs, slog.Any("error_chain", chain))
		}
	}

	return attrs
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
