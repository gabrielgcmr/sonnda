package observability

import (
	"context"
	"log/slog"
)

// IntoContext coloca um *slog.Logger no context.Context.
func IntoContext(ctx context.Context, l *slog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if l == nil {
		l = slog.Default()
	}
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext obtém o logger do contexto; se não existir, retorna slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return slog.Default()
}
