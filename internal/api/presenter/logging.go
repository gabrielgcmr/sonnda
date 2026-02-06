// internal/api/presenter/logging.go
package presenter

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/gin-gonic/gin"
)

func buildLogAttrs(c *gin.Context, status int, code apperr.ErrorKind, err error) []any {
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
	seen := map[string]struct{}{}
	for err != nil {
		key := errorKey(err)
		if _, ok := seen[key]; ok {
			chain = append(chain, "cycle_detected")
			break
		}
		seen[key] = struct{}{}

		chain = append(chain, describeError(err))
		err = errors.Unwrap(err)
	}

	return chain
}

func describeError(err error) string {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil {
		return fmt.Sprintf("app_error code=%s msg=%s", appErr.Kind, appErr.Message)
	}
	return err.Error()
}

func errorKey(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%T:%s", err, err.Error())
}
