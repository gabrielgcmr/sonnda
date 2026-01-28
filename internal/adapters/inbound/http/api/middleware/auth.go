package middleware

import (
	"strings"

	sharedauth "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/auth"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/httpctx"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/httperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/security"

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
		// 1) Extrai Authorization
		h := strings.TrimSpace(c.GetHeader("Authorization"))
		if h == "" {
			httperr.APIErrorResponder(c, apperr.Unauthorized("missing authorization header"))
			c.Abort()
			return
		}

		// 2) Extrai token Bearer
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httperr.APIErrorResponder(c, apperr.Unauthorized("invalid authorization header"))
			c.Abort()
			return
		}
		token := strings.TrimSpace(parts[1])
		if token == "" {
			httperr.APIErrorResponder(c, apperr.Unauthorized("missing bearer token"))
			c.Abort()
			return
		}

		// 3) Autentica
		id, err := m.core.AuthenticateFromBearerToken(c.Request.Context(), token)
		if err != nil {
			// 4) Responde no formato padrão
			if ae, ok := err.(*apperr.AppError); ok {
				httperr.APIErrorResponder(c, ae)
			} else {
				httperr.APIErrorResponder(c, apperr.Internal("internal auth error", err))
			}
			c.Abort()
			return
		}

		// 5) Injeta Identity no contexto e segue
		httpctx.SetIdentity(c, (*security.Identity)(id))
		c.Next()
	}
}
