// internal/adapters/inbound/http/api/middleware/registration.go
package middleware

import (
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/apierr"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/httpctx"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/register"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/gin-gonic/gin"
)

// RegistrationMiddleware (API)
// Responsável por:
// - garantir que exista um User local associado à Identity
// - responder com JSON (contrato da API) quando não existir
type RegistrationMiddleware struct {
	core *register.Core
}

func NewRegistrationMiddleware(core *register.Core) *RegistrationMiddleware {
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
		id, ok := httpctx.GetIdentity(c)
		if !ok || id == nil {
			// Situação anômala: auth não rodou antes
			apierr.ErrorResponder(c, &apperr.AppError{
				Code:    apperr.AUTH_REQUIRED,
				Message: "autenticação necessária",
			})
			return
		}

		u, appErr := m.core.ResolveCurrentUser(c.Request.Context(), id)
		if appErr != nil {
			apierr.ErrorResponder(c, appErr)
			return
		}

		if u == nil {
			// Autenticado no provider, mas sem cadastro local
			apierr.ErrorResponder(c, &apperr.AppError{
				Code:    apperr.ACCESS_DENIED,
				Message: "cadastro necessário",
			})
			return
		}

		httpctx.SetCurrentUser(c, u)
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
