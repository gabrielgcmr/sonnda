//internal/app/config/observability/logger.go

package observability

import (
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
	switch strings.ToLower(cfg.Format) {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, opts)
	case "pretty":
		h = NewPrettyHandler(os.Stdout, opts.Level)
	default:
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
