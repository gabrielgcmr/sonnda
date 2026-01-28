// internal/adapters/inbound/http/shared/registration/core.go
package register

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/model/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

// Core: resolve o User (se existir) mas NÃO responde HTTP.
type Core struct {
	userRepo ports.UserRepo
}

func NewCore(userRepo ports.UserRepo) *Core {
	return &Core{userRepo: userRepo}
}

// ResolveCurrentUser tenta encontrar o usuário cadastrado no seu banco.
// - Se não existir: retorna (nil, nil) para o adapter decidir (api=JSON, web=redirect).
// - Se erro infra: retorna AppError internal/infra conforme você preferir.
func (c *Core) ResolveCurrentUser(ctx context.Context, id *security.Identity) (*user.User, error) {
	if id == nil {
		return nil, apperr.Unauthorized("autenticação necessária")
	}

	u, err := c.userRepo.FindByPrincipalID(ctx, id.PrincipalID())
	if err != nil {
		return nil, apperr.Internal("falha ao buscar usuário", err)
	}

	// nil significa: autenticado no provider, mas ainda não existe cadastro local.
	return u, nil
}
