// internal/api/humaerror/observability.go
package humaerror

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	applog "github.com/gabrielgcmr/sonnda/internal/kernel/observability"
)

// Transform records handler failures at Huma's response boundary and hands only
// the standard public model to the remaining transformers and serializer.
// Register it before schema transformers on the application's Gin-backed API.
func Transform(ctx huma.Context, _ string, body any) (any, error) {
	if err, ok := body.(*Error); ok {
		observe(ctx, err)
		return err.ErrorModel, nil
	}
	return body, nil
}

func observe(ctx huma.Context, err *Error) {
	c := humagin.Unwrap(ctx)
	c.Set("error_code", string(err.code))
	c.Set("error_log_level", apperr.LogLevelOf(err))
	if err.Status < 500 || c.GetBool("huma_error_logged") || c.GetBool("panic_recovered") {
		return
	}
	c.Set("huma_error_logged", true)
	applog.FromContext(ctx.Context()).ErrorContext(ctx.Context(), "handler_error",
		slog.Int("status", err.Status),
		slog.String("error_code", string(err.code)),
		slog.String("route", c.FullPath()),
		slog.Any("err", err.cause),
		slog.Any("error_chain", errorChain(err.cause)),
	)
}

// Include joined errors as well as ordinary wrappers. Bound traversal so a
// malformed or cyclic error cannot prevent the HTTP response from being sent.
func errorChain(err error) []string {
	var chain []string
	var visit func(error)
	visit = func(current error) {
		if current == nil || len(chain) >= 64 {
			return
		}
		chain = append(chain, current.Error())
		switch wrapped := current.(type) {
		case interface{ Unwrap() []error }:
			for _, child := range wrapped.Unwrap() {
				visit(child)
			}
		case interface{ Unwrap() error }:
			visit(wrapped.Unwrap())
		}
	}
	visit(err)
	if len(chain) == 64 {
		chain = append(chain, "error_chain_truncated")
	}
	return chain
}
