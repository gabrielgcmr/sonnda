// internal/features/account/http/middleware.go
package accounthttp

import (
	"context"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	"github.com/gabrielgcmr/sonnda/internal/features/account"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	authdomain "github.com/gabrielgcmr/sonnda/internal/features/auth/domain"
	authhttp "github.com/gabrielgcmr/sonnda/internal/features/auth/http"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	userRepo account.Repository
}

func NewMiddleware(userRepo account.Repository) *Middleware {
	return &Middleware{userRepo: userRepo}
}

func (m *Middleware) resolveCurrentUser(ctx context.Context, identity *authdomain.Identity) (*accountdomain.User, error) {
	if identity == nil {
		return nil, apperr.Unauthorized("autenticação necessária")
	}

	currentUser, err := m.userRepo.FindByAuthIdentity(ctx, identity.Issuer, identity.Subject)
	if err != nil {
		return nil, apperr.Internal("falha ao buscar usuário", err)
	}

	return currentUser, nil
}

func (m *Middleware) RequireRegisteredUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, ok := authhttp.GetIdentity(c)
		if !ok || identity == nil {
			presenter.ErrorResponder(c, apperr.Unauthorized("autenticação necessária"))
			return
		}

		currentUser, err := m.resolveCurrentUser(c.Request.Context(), identity)
		if err != nil {
			presenter.ErrorResponder(c, err)
			return
		}
		if currentUser == nil {
			presenter.ErrorResponder(c, apperr.ProfileNotFound())
			return
		}

		helpers.SetCurrentUser(c, currentUser)
		c.Next()
	}
}

func (m *Middleware) LoadCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, ok := authhttp.GetIdentity(c)
		if !ok || identity == nil {
			c.Next()
			return
		}

		currentUser, _ := m.resolveCurrentUser(c.Request.Context(), identity)
		if currentUser != nil {
			helpers.SetCurrentUser(c, currentUser)
		}

		c.Next()
	}
}
