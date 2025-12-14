package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
)

type ctxKey struct{}

var key ctxKey

type Config struct {
	Env       string // "dev" | "prod"
	Level     slog.Level
	AppName   string
	AddSource bool
}

// New cria um *slog.Logger pronto pra uso.
func New(cfg Config) *slog.Logger {
	// Defaults
	if cfg.Env == "" {
		cfg.Env = "dev"
	}
	if cfg.AppName == "" {
		cfg.AppName = "sonnda-api"
	}

	var h slog.Handler
	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Deixa timestamp no padrão ISO
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.UTC().Format(time.RFC3339Nano))
				}
			}
			return a
		},
	}

	// prod -> JSON, dev -> text (mais legível)
	if strings.EqualFold(cfg.Env, "prod") {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	l := slog.New(h).With(
		slog.String("app", cfg.AppName),
		slog.String("env", cfg.Env),
	)
	return l
}

func IntoContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, key, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if v := ctx.Value(key); v != nil {
		if l, ok := v.(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return slog.Default()
}
