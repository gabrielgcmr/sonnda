package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

//Error: falha do servidor / bug / dependência fora / banco caiu / docAI falhou
//Warn: algo “estranho” que pode indicar problema, abuso ou mau uso (ex.: rate-limit, tentativa inválida repetida, payload malformado demais)
//Info: comportamento esperado (inclui muitos 4xx)
//Debug: detalhes úteis pra dev

type Config struct {
	Env       string // "dev" | "prod"
	Level     string
	Format    string // "text" | "json"
	AppName   string
	AddSource bool
}

// chave privada pra evitar colisão com outros packages
type ctxKey struct{}

var loggerKey ctxKey

// New cria um *slog.Logger pronto pra uso.
func New(cfg Config) *slog.Logger {
	// Defaults
	if cfg.AppName == "" {
		cfg.AppName = "sonnda-api"
	}

	level := parseLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     level,
	}

	// prod -> JSON, dev -> text (mais legível)
	var h slog.Handler
	if strings.EqualFold(cfg.Format, "json") {
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

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

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
