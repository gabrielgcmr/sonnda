// internal/adapters/inbound/http/web/middleware/auth.go
package middleware

import (
	"net/http"

	"sonnda-api/internal/adapters/inbound/http/shared/auth"
	"sonnda-api/internal/adapters/inbound/http/shared/httpctx"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware (WEB) exige autenticação via cookie de sessão (__session).
// - Resolve Identity via shared/auth Core (sem acoplamento ao Gin).
// - Em caso de erro, faz redirect (comportamento de browser).
type AuthMiddleware struct {
	core *auth.Core
}

func NewAuthMiddleware(core *auth.Core) *AuthMiddleware {
	return &AuthMiddleware{core: core}
}

// RequireSession:
// - Lê cookie __session
// - Verifica sessão no provider (IdentityService)
// - Coloca Identity no contexto da requisição (reqctx)
// - Caso falhe: redirect para /login (por enquanto) e aborta.
func (m *AuthMiddleware) RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, appErr := m.core.AuthenticateFromSessionCookie(c.Request.Context(), c.Request)
		if appErr != nil {
			// TODO: você pode melhorar isso depois adicionando ?next=/rota-atual
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		httpctx.SetIdentity(c, id)
		c.Next()
	}
}
