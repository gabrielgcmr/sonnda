// internal/adapters/inbound/http/shared/registration/core.go
package register

import (
	"context"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/identity"
	"sonnda-api/internal/domain/model/user"
)

// UserResolver define a interface necessária para resolver Users a partir de Identity.
type UserResolver interface {
	FindByAuthIdentity(ctx context.Context, provider, subject string) (*user.User, error)
}

// Core: resolve o User (se existir) mas NÃO responde HTTP.
type Core struct {
	users UserResolver
}

func NewCore(users UserResolver) *Core {
	return &Core{users: users}
}

// ResolveCurrentUser tenta encontrar o usuário cadastrado no seu banco.
// - Se não existir: retorna (nil, nil) para o adapter decidir (api=JSON, web=redirect).
// - Se erro infra: retorna AppError internal/infra conforme você preferir.
func (c *Core) ResolveCurrentUser(ctx context.Context, id *identity.Identity) (*user.User, *apperr.AppError) {
	if id == nil {
		return nil, &apperr.AppError{
			Code:    apperr.AUTH_REQUIRED,
			Message: "autenticação necessária",
		}
	}

	u, err := c.users.FindByAuthIdentity(ctx, id.Provider, id.Subject)
	if err != nil {
		return nil, &apperr.AppError{
			Code:    apperr.INFRA_DATABASE_ERROR,
			Message: "erro ao carregar usuário",
			Cause:   err,
		}
	}

	// nil significa: autenticado no provider, mas ainda não existe cadastro local.
	return u, nil
}
