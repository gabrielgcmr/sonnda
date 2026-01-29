// internal/api/middleware/auth.go
package middleware

import (
	"context"
	"strings"

	apictx "github.com/gabrielgcmr/sonnda/internal/api/apictx"
	"github.com/gabrielgcmr/sonnda/internal/api/apierr"
	"github.com/gabrielgcmr/sonnda/internal/domain/model"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware (API) exige autenticação via Bearer token.
// - Resolve Identity via função de autenticação (sem acoplamento ao Gin).
// - Em caso de erro, escreve JSON usando o contrato da API (apierr).
type AuthMiddleware struct {
	authenticate func(context.Context, string) (*model.Identity, error)
}

func NewAuthMiddleware(authenticate func(context.Context, string) (*model.Identity, error)) *AuthMiddleware {
	return &AuthMiddleware{authenticate: authenticate}
}

// RequireBearer:
// - Lê Authorization: Bearer <token>
// - Verifica token no provider
// - Coloca Identity no contexto da requisição (reqctx)
// - Caso falhe: responde JSON (apierr) e aborta.
func (m *AuthMiddleware) RequireBearer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Extrai Authorization
		h := strings.TrimSpace(c.GetHeader("Authorization"))
		if h == "" {
			apierr.ErrorResponder(c, apperr.Unauthorized("missing authorization header"))
			c.Abort()
			return
		}

		// 2) Extrai token Bearer
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			apierr.ErrorResponder(c, apperr.Unauthorized("invalid authorization header"))
			c.Abort()
			return
		}
		token := strings.TrimSpace(parts[1])
		if token == "" {
			apierr.ErrorResponder(c, apperr.Unauthorized("missing bearer token"))
			c.Abort()
			return
		}

		// 3) Autentica
		id, err := m.authenticate(c.Request.Context(), token)
		if err != nil {
			// 4) Responde no formato padrão
			if ae, ok := err.(*apperr.AppError); ok {
				apierr.ErrorResponder(c, ae)
			} else {
				apierr.ErrorResponder(c, apperr.Internal("internal auth error", err))
			}
			c.Abort()
			return
		}
		if id == nil {
			apierr.ErrorResponder(c, apperr.Unauthorized("token inválido ou expirado"))
			c.Abort()
			return
		}

		// 5) Injeta Identity no contexto e segue
		apictx.SetIdentity(c, id)
		c.Next()
	}
}
