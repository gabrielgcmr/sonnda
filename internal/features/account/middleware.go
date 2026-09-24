// internal/features/account/middleware.go
package account

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/identity"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	userRepo repository.User
}

func NewMiddleware(userRepo repository.User) *Middleware {
	return &Middleware{userRepo: userRepo}
}

func (m *Middleware) resolveCurrentUser(ctx context.Context, identity *identity.Identity) (*user.User, error) {
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
		identity, ok := helpers.GetIdentity(c)
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
		identity, ok := helpers.GetIdentity(c)
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
