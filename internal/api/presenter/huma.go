// internal/api/presenter/huma.go
package presenter

import (
	"context"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/observability"
	"github.com/gin-gonic/gin"
)

type errorContextKey struct{}

func WithErrorContext(ctx context.Context, c *gin.Context) context.Context {
	return context.WithValue(ctx, errorContextKey{}, c)
}
func HumaError(ctx context.Context, err error) error {
	var app *apperr.AppError
	if !errors.As(err, &app) {
		app = apperr.Internal("Erro inesperado.", err)
	}
	status := StatusFromCode(app.Kind)
	if c, ok := ctx.Value(errorContextKey{}).(*gin.Context); ok {
		c.Set("error_code", string(app.Kind))
		_ = c.Error(err)
	}
	if status >= 500 {
		observability.FromContext(ctx).ErrorContext(ctx, "handler_error", "error_code", app.Kind, "error_chain", errorChain(err))
	}
	return huma.NewError(status, app.Message)
}
