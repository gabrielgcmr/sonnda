package middleware

import (
	"sonnda-api/internal/adapters/inbound/http/api/httperr"
	sharedauth "sonnda-api/internal/adapters/inbound/http/shared/auth"
	"sonnda-api/internal/adapters/inbound/http/shared/reqctx"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware (API) exige autenticação via Bearer token.
// - Resolve Identity via shared/auth Core (sem acoplamento ao Gin).
// - Em caso de erro, escreve JSON usando o contrato da API (httperr).
type AuthMiddleware struct {
	core *sharedauth.Core
}

func NewAuthMiddleware(core *sharedauth.Core) *AuthMiddleware {
	return &AuthMiddleware{core: core}
}

// RequireBearer:
// - Lê Authorization: Bearer <token>
// - Verifica token no provider (IdentityService)
// - Coloca Identity no contexto da requisição (reqctx)
// - Caso falhe: responde JSON (httperr) e aborta.
func (m *AuthMiddleware) RequireBearer() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, appErr := m.core.AuthenticateFromBearerToken(c.Request.Context(), c.Request)
		if appErr != nil {
			httperr.WriteError(c, appErr)
			return
		}

		reqctx.SetIdentity(c, id)
		c.Next()
	}
}
