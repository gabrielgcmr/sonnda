// internal/api/apierr/presenter.go
package presenter

import (
	"log/slog"

	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	applog "github.com/gabrielgcmr/sonnda/internal/kernel/observability"

	"github.com/gin-gonic/gin"
)

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

	// Mapeia erro para HTTP + contrato público.
	status, resp := ToHTTP(err)
	level := apperr.LogLevelOf(err)

	// Metadados para observabilidade.
	c.Set("error_code", string(resp.Code))
	c.Set("http_status", status)
	c.Set("error_log_level", level)

	log := applog.FromContext(c.Request.Context())

	// Log estruturado apenas quando necessário.
	attrs := buildLogAttrs(c, status, resp.Code, err)
	if level >= slog.LevelWarn || status >= 500 {
		log.Log(c.Request.Context(), level, "handler_error", attrs...)
	}

	// Resposta final padronizada.
	c.Abort()
	c.JSON(status, gin.H{"error": resp})
}
