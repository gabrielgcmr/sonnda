// internal/api/middleware/registration.go
package middleware

import (
	"context"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	"github.com/gabrielgcmr/sonnda/internal/domain/model"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/gin-gonic/gin"
)

// RegistrationMiddleware (API)
// Responsável por:
// - garantir que exista um User local associado à Identity
// - responder com JSON (contrato da API) quando não existir
type RegistrationMiddleware struct {
	userRepo ports.UserRepo
}

func NewRegistrationMiddleware(userRepo ports.UserRepo) *RegistrationMiddleware {
	return &RegistrationMiddleware{userRepo: userRepo}
}

// resolveCurrentUser tenta encontrar o usuário cadastrado no seu banco.
// - Se não existir: retorna (nil, nil) para o adapter decidir (api=JSON).
// - Se erro infra: retorna AppError internal/infra conforme você preferir.
func (m *RegistrationMiddleware) resolveCurrentUser(ctx context.Context, id *model.Identity) (*user.User, error) {
	if id == nil {
		return nil, apperr.Unauthorized("autenticação necessária")
	}

	u, err := m.userRepo.FindByAuthIdentity(ctx, id.Issuer, id.Subject)
	if err != nil {
		return nil, apperr.Internal("falha ao buscar usuário", err)
	}

	return u, nil
}

// RequireRegisteredUser
// Fluxo típico da API:
//
// 1) Identity JÁ DEVE estar no contexto (auth middleware roda antes)
// 2) Resolve User local a partir da Identity
// 3) Se não existir → erro JSON (ex: 403 / cadastro necessário)
// 4) Se existir → coloca CurrentUser no contexto
func (m *RegistrationMiddleware) RequireRegisteredUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := helpers.GetIdentity(c)
		if !ok || id == nil {
			// Situação anômala: auth não rodou antes
			presenter.ErrorResponder(c, apperr.Unauthorized("autenticação necessária"))
			return
		}

		u, err := m.resolveCurrentUser(c.Request.Context(), id)
		if err != nil {
			presenter.ErrorResponder(c, err)
			return
		}

		if u == nil {
			// Autenticado no provider, mas sem cadastro local
			presenter.ErrorResponder(c, apperr.Forbidden("cadastro necessário"))
			return
		}

		helpers.SetCurrentUser(c, u)
		c.Next()
	}
}

// LoadCurrentUser
// Variante "best effort":
// - tenta carregar o User
// - NÃO bloqueia se não existir
// Útil para endpoints que aceitam usuário anônimo + autenticado
func (m *RegistrationMiddleware) LoadCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := helpers.GetIdentity(c)
		if !ok || id == nil {
			c.Next()
			return
		}

		u, _ := m.resolveCurrentUser(c.Request.Context(), id)
		if u != nil {
			helpers.SetCurrentUser(c, u)
		}

		c.Next()
	}
}
