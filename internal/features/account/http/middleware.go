// internal/features/account/http/middleware.go
package accounthttp

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/features/account"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	authdomain "github.com/gabrielgcmr/sonnda/internal/features/auth/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type Middleware struct {
	userRepo account.Repository
}

func NewMiddleware(userRepo account.Repository) *Middleware {
	return &Middleware{userRepo: userRepo}
}

func (m *Middleware) ResolveRegisteredUser(ctx context.Context, identity *authdomain.Identity) (*accountdomain.User, error) {
	if identity == nil {
		return nil, apperr.Unauthorized("autenticação necessária")
	}

	currentUser, err := m.userRepo.FindByAuthIdentity(ctx, identity.Issuer, identity.Subject)
	if err != nil {
		return nil, apperr.Internal("falha ao buscar usuário", err)
	}

	return currentUser, nil
}
