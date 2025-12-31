// internal/app/config/logger/pretty_handler.go
package observability

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

/*
ANSI COLORS
*/
const (
	reset = "\033[0m"
	bold  = "\033[1m"

	gray   = "\033[90m"
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
)

type PrettyHandler struct {
	out   io.Writer
	level slog.Leveler
}

func NewPrettyHandler(out io.Writer, level slog.Leveler) *PrettyHandler {
	return &PrettyHandler{
		out:   out,
		level: level,
	}
}

func (h *PrettyHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level.Level()
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	// ───── header ──────────────────────────────────────────────
	ts := r.Time.Format("15:04:05")

	levelColor := colorForLevel(r.Level)
	levelLabel := fmt.Sprintf("%s%s%-5s%s", bold, levelColor, r.Level.String(), reset)

	method, path := "", ""
	attrs := make([]slog.Attr, 0, 8)

	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "method":
			method = a.Value.String()
		case "path":
			path = a.Value.String()
		default:
			attrs = append(attrs, a)
		}
		return true
	})

	title := r.Message
	if method != "" && path != "" {
		title = fmt.Sprintf("%s %s %s", title, method, path)
	}

	fmt.Fprintf(h.out, "%s%s%s %s %s\n",
		gray, ts, reset,
		levelLabel,
		bold+title+reset,
	)

	fmt.Fprintln(h.out, gray+strings.Repeat("─", 72)+reset)

	// ───── attributes ──────────────────────────────────────────
	for _, a := range attrs {
		if a.Key == "stack" {
			h.printStack(a.Value.String())
			continue
		}

		fmt.Fprintf(
			h.out,
			"  %s%-12s%s : %s\n",
			cyan,
			a.Key,
			reset,
			fmt.Sprintf("%v", a.Value.Any()),
		)
	}

	fmt.Fprintln(h.out)
	return nil
}

func (h *PrettyHandler) printStack(stack string) {
	fmt.Fprintf(h.out, "\n  %sSTACK TRACE%s\n", red, reset)
	fmt.Fprintln(h.out, "  "+gray+strings.Repeat("─", 68)+reset)

	for _, line := range strings.Split(stack, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fmt.Fprintf(h.out, "    %s\n", line)
	}

	fmt.Fprintln(h.out, "  "+gray+strings.Repeat("─", 68)+reset)
}

func (h *PrettyHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *PrettyHandler) WithGroup(_ string) slog.Handler {
	return h
}

func colorForLevel(l slog.Level) string {
	switch l {
	case slog.LevelDebug:
		return cyan
	case slog.LevelInfo:
		return green
	case slog.LevelWarn:
		return yellow
	case slog.LevelError:
		return red
	default:
		return blue
	}
}
