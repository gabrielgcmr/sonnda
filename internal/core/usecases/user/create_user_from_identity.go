// internal/core/usecases/user/create_user_from_identity.go
package user

import (
	"context"
	"time"

	"sonnda-api/internal/core/domain/identity"
	"sonnda-api/internal/core/ports/repositories"
	"sonnda-api/internal/core/ports/services"
)

type CreateUserFromIdentity struct {
	userRepo repositories.UserRepository
}

func NewCreateUserFromIdentity(userRepo repositories.UserRepository) *CreateUserFromIdentity {
	return &CreateUserFromIdentity{
		userRepo: userRepo,
	}
}

// Execute cria um app_users com RolePatient para a Identity informada,
// caso ainda não exista. Se já existir, apenas retorna.
func (uc *CreateUserFromIdentity) Execute(
	ctx context.Context,
	i *services.Identity,
) (*identity.User, error) {
	//1. Normaliza o provider antes de buscar
	provider := i.Provider
	if provider == "" {
		provider = "firebase"
	}

	// 2. Verifica se já existe
	existing, err := uc.userRepo.FindByAuthIdentity(
		ctx, provider,
		i.Subject)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	// 2.1. Se já existe um usuário com o mesmo email, reusa e atualiza a identidade.
	existingByEmail, err := uc.userRepo.FindByEmail(ctx, i.Email)
	if err != nil {
		return nil, err
	}
	if existingByEmail != nil {
		// Atualiza provider/subject para o valor atual e retorna.
		updated, err := uc.userRepo.UpdateAuthIdentity(
			ctx, existingByEmail.ID, provider, i.Subject)
		if err != nil {
			return nil, err
		}
		return updated, nil
	}

	// 2. Cria novo
	now := time.Now()
	u := &identity.User{
		AuthProvider: provider,
		AuthSubject:  i.Subject,
		Email:        i.Email,
		Role:         identity.RolePatient, // paciente se auto-registra
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}
