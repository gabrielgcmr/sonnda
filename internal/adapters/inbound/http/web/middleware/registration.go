// internal/adapters/inbound/http/web/middleware/registration.go
package middleware

import (
	"net/http"

	"sonnda-api/internal/adapters/inbound/http/shared/httpctx"
	"sonnda-api/internal/adapters/inbound/http/shared/registration"

	"github.com/gin-gonic/gin"
)

// RegistrationMiddleware (WEB)
// Responsável por:
// - garantir que exista um User local associado à Identity
// - redirecionar quando não existir (UX de browser)
type RegistrationMiddleware struct {
	core *registration.Core
}

func NewRegistrationMiddleware(core *registration.Core) *RegistrationMiddleware {
	return &RegistrationMiddleware{core: core}
}

// RequireRegisteredUser
// Fluxo típico do WEB:
//
// 1) Identity JÁ DEVE estar no contexto (auth de sessão rodou antes)
// 2) Resolve User local
// 3) Se não existir → redirect para onboarding
// 4) Se existir → coloca CurrentUser no contexto
func (m *RegistrationMiddleware) RequireRegisteredUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := httpctx.GetIdentity(c)
		if !ok || id == nil {
			// Não autenticado → login
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		u, _ := m.core.ResolveCurrentUser(c.Request.Context(), id)
		if u == nil {
			// Autenticado, mas sem cadastro local
			c.Redirect(http.StatusFound, "/onboarding")
			c.Abort()
			return
		}

		httpctx.SetCurrentUser(c, u)
		c.Next()
	}
}

// LoadCurrentUser
// Best effort no WEB também:
// - útil para header (mostrar nome se logado, etc.)
func (m *RegistrationMiddleware) LoadCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := httpctx.GetIdentity(c)
		if !ok || id == nil {
			c.Next()
			return
		}

		u, _ := m.core.ResolveCurrentUser(c.Request.Context(), id)
		if u != nil {
			httpctx.SetCurrentUser(c, u)
		}

		c.Next()
	}
}
