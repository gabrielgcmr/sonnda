// internal/adapters/inbound/http/shared/httperr/base.go
package httperr

import (
	"errors"
	"log/slog"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	applog "github.com/gabrielgcmr/sonnda/internal/kernel/observability"
	"github.com/gin-gonic/gin"
)

// ErrorPresenter define como um erro é apresentado ao cliente
type ErrorPresenter interface {
	Present(c *gin.Context, status int, resp ErrorResponse)
}

// BaseErrorResponder contém lógica compartilhada entre API e Web
func BaseErrorResponder(c *gin.Context, err error, presenter ErrorPresenter) {
	// Evita responder duas vezes
	if c.Writer.Written() {
		c.Abort()
		return
	}

	// Registra erro no Gin context (útil para middleware de log)
	if err != nil {
		_ = c.Error(err)
	}

	// Mapeia erro para HTTP
	status, resp := ToHTTP(err)
	level := apperr.LogLevelOf(err)

	// Metadados para observabilidade
	c.Set("error_code", string(resp.Code))
	c.Set("http_status", status)
	c.Set("error_log_level", level)

	// Log estruturado
	logError(c, err, status, resp.Code, level)

	// Delega apresentação para o presenter específico
	c.Abort()
	presenter.Present(c, status, resp)
}

func logError(c *gin.Context, err error, status int, code apperr.ErrorCode, level slog.Level) {
	log := applog.FromContext(c.Request.Context())

	attrs := []any{
		slog.Int("status", status),
		slog.String("error_code", string(code)),
		slog.String("path", c.FullPath()),
		slog.String("method", c.Request.Method),
	}

	if err != nil {
		attrs = append(attrs, slog.Any("err", err))
		if chain := buildErrorChain(err); len(chain) > 0 {
			attrs = append(attrs, slog.Any("error_chain", chain))
		}
	}

	// Log apenas erros relevantes
	if level >= slog.LevelWarn || status >= 500 {
		log.Log(c.Request.Context(), level, "handler_error", attrs...)
	}
}

func buildErrorChain(err error) []string {
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
		return "app_error code=" + string(appErr.Code) + " msg=" + appErr.Message
	}
	return err.Error()
}
