// internal/features/auth/middleware.go
package auth

import (
	"context"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/identity"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	authenticate func(context.Context, string) (*identity.Identity, error)
}

func NewMiddleware(authenticate func(context.Context, string) (*identity.Identity, error)) *Middleware {
	return &Middleware{authenticate: authenticate}
}

func (m *Middleware) RequireBearer() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			presenter.ErrorResponder(c, apperr.Unauthorized("missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			presenter.ErrorResponder(c, apperr.Unauthorized("invalid authorization header"))
			c.Abort()
			return
		}
		token := strings.TrimSpace(parts[1])
		if token == "" {
			presenter.ErrorResponder(c, apperr.Unauthorized("missing bearer token"))
			c.Abort()
			return
		}

		identity, err := m.authenticate(c.Request.Context(), token)
		if err != nil {
			if appErr, ok := err.(*apperr.AppError); ok {
				presenter.ErrorResponder(c, appErr)
			} else {
				presenter.ErrorResponder(c, apperr.Internal("internal auth error", err))
			}
			c.Abort()
			return
		}
		if identity == nil {
			presenter.ErrorResponder(c, apperr.Unauthorized("token inválido ou expirado"))
			c.Abort()
			return
		}

		helpers.SetIdentity(c, identity)
		c.Next()
	}
}
