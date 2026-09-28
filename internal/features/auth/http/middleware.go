// internal/features/auth/http/middleware.go
package authhttp

import (
	"context"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	authdomain "github.com/gabrielgcmr/sonnda/internal/features/auth/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	authenticate func(context.Context, string) (*authdomain.Identity, error)
}

func NewMiddleware(authenticate func(context.Context, string) (*authdomain.Identity, error)) *Middleware {
	return &Middleware{authenticate: authenticate}
}

// AuthenticateBearer validates an Authorization header independently of the
// HTTP adapter, so both Gin and Huma use the same authentication policy.
func (m *Middleware) AuthenticateBearer(ctx context.Context, authorization string) (*authdomain.Identity, error) {
	header := strings.TrimSpace(authorization)
	if header == "" {
		return nil, apperr.Unauthorized("missing authorization header")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, apperr.Unauthorized("invalid authorization header")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return nil, apperr.Unauthorized("missing bearer token")
	}

	identity, err := m.authenticate(ctx, token)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return nil, appErr
		}
		return nil, apperr.Internal("internal auth error", err)
	}
	if identity == nil {
		return nil, apperr.Unauthorized("token inválido ou expirado")
	}

	return identity, nil
}

func (m *Middleware) RequireBearer() gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, err := m.AuthenticateBearer(c.Request.Context(), c.GetHeader("Authorization"))
		if err != nil {
			presenter.ErrorResponder(c, err)
			c.Abort()
			return
		}

		SetIdentity(c, identity)
		c.Next()
	}
}
