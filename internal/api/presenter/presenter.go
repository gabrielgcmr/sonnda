// internal/api/presenter/presenter.go
package presenter

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/api/problem"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	applog "github.com/gabrielgcmr/sonnda/internal/kernel/observability"

	"github.com/gin-gonic/gin"
)

const problemJSONContentType = "application/problem+json; charset=utf-8"

func ErrorResponder(c *gin.Context, err error) {
	// Evita responder duas vezes.
	if c.Writer.Written() {
		c.Abort()
		return
	}
	// Registra erro no contexto do Gin para middleware de log.
	if err != nil {
		_ = c.Error(err)
	}

	rid := strings.TrimSpace(c.GetString("request_id"))
	instance := ""
	if rid != "" {
		instance = "urn:sonnda:request-id:" + rid
	} else if c.Request != nil && c.Request.URL != nil {
		instance = c.Request.URL.Path
	}

	// Mapeia erro para HTTP + contrato público (RFC 9457).
	status, resp := ToProblem(err, problem.Meta{
		Instance:  instance,
		RequestID: rid,
	})
	level := apperr.LogLevelOf(err)

	// Metadados para observabilidade.
	c.Set("error_code", string(resp.Code))
	c.Set("http_status", status)
	c.Set("error_log_level", level)

	log := applog.FromContext(c.Request.Context())

	// Log estruturado apenas quando necessário.
	attrs := buildLogAttrs(c, status, resp.Code, err)
	if !shouldSkipErrorLog(c) && (level >= slog.LevelWarn || status >= 500) {
		log.Log(c.Request.Context(), level, "handler_error", attrs...)
	}

	// Resposta final padronizada.
	c.Abort()
	writeProblem(c, status, resp)
}

func shouldSkipErrorLog(c *gin.Context) bool {
	if c == nil {
		return false
	}
	v, ok := c.Get("panic_recovered")
	if !ok {
		return false
	}
	skip, _ := v.(bool)
	return skip
}

func writeProblem(c *gin.Context, status int, body problem.Details) {
	if c == nil {
		return
	}
	b, err := json.Marshal(body)
	if err != nil {
		// Falha de marshal não deve acontecer; fallback seguro.
		b = []byte(`{"type":"about:blank","title":"Erro","status":500,"detail":"erro inesperado","code":"INTERNAL_ERROR"}`)
		status = 500
	}

	c.Writer.Header().Set("Content-Type", problemJSONContentType)
	c.Writer.WriteHeader(status)
	_, _ = c.Writer.Write(b)
}
