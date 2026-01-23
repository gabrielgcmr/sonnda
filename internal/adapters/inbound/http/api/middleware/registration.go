// internal/adapters/inbound/http/api/middleware/registration.go
package middleware

import (
	"sonnda-api/internal/adapters/inbound/http/api/httperr"
	"sonnda-api/internal/adapters/inbound/http/shared/registration"
	"sonnda-api/internal/adapters/inbound/http/shared/reqctx"
	"sonnda-api/internal/app/apperr"

	"github.com/gin-gonic/gin"
)

// RegistrationMiddleware (API)
// Responsável por:
// - garantir que exista um User local associado à Identity
// - responder com JSON (contrato da API) quando não existir
type RegistrationMiddleware struct {
	core *registration.Core
}

func NewRegistrationMiddleware(core *registration.Core) *RegistrationMiddleware {
	return &RegistrationMiddleware{core: core}
}

// RequireRegisteredUser
// Fluxo típico da API:
//
// 1) Identity JÁ DEVE estar no contexto (auth middleware roda antes)
// 2) Resolve User local a partir da Identity
// 3) Se não existir → erro JSON (ex: 403 / cadastro necessário)
// 4) Se existir → coloca CurrentUser no contexto
func (m *RegistrationMiddleware) RequireRegisteredUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := reqctx.GetIdentity(c)
		if !ok || id == nil {
			// Situação anômala: auth não rodou antes
			httperr.WriteError(c, &apperr.AppError{
				Code:    apperr.AUTH_REQUIRED,
				Message: "autenticação necessária",
			})
			return
		}

		u, appErr := m.core.ResolveCurrentUser(c.Request.Context(), id)
		if appErr != nil {
			httperr.WriteError(c, appErr)
			return
		}

		if u == nil {
			// Autenticado no provider, mas sem cadastro local
			httperr.WriteError(c, &apperr.AppError{
				Code:    apperr.ACCESS_DENIED,
				Message: "cadastro necessário",
			})
			return
		}

		reqctx.SetCurrentUser(c, u)
		c.Next()
	}
}

// LoadCurrentUser
// Variante "best effort":
// - tenta carregar o User
// - NÃO bloqueia se não existir
// Útil para endpoints que aceitam usuário anônimo + autenticado
func (m *RegistrationMiddleware) LoadCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := reqctx.GetIdentity(c)
		if !ok || id == nil {
			c.Next()
			return
		}

		u, _ := m.core.ResolveCurrentUser(c.Request.Context(), id)
		if u != nil {
			reqctx.SetCurrentUser(c, u)
		}

		c.Next()
	}
}
