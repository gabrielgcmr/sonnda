// internal/adapters/outbound/database/supabase/user_repository.go
package supabase

import (
	"context"
	"errors"
	"fmt"

	userssqlc "sonnda-api/internal/adapters/outbound/database/sqlc/users"

	"sonnda-api/internal/core/domain/identity"
	"sonnda-api/internal/core/ports/repositories"

	"github.com/jackc/pgx/v5"
)

var _ repositories.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	client  *Client
	queries *userssqlc.Queries
}

func NewUserRepository(client *Client) *UserRepository {
	return &UserRepository{
		client:  client,
		queries: userssqlc.New(client.Pool()),
	}
}

func (r *UserRepository) FindByAuthIdentity(
	ctx context.Context,
	provider, subject string,
) (*identity.User, error) {
	dbUser, err := r.queries.FindUserByAuthIdentity(ctx, userssqlc.FindUserByAuthIdentityParams{
		AuthProvider: provider,
		AuthSubject:  subject,
	})
	if err != nil {
		// IMPORTANTE: tratar "nao encontrado" como (nil, nil)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return dbUserToDomain(dbUser)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	dbUser, err := r.queries.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return dbUserToDomain(dbUser)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*identity.User, error) {
	dbUser, err := r.queries.FindUserByID(ctx, ToPgUUID(id))
	if err != nil {
		return nil, err
	}

	return dbUserToDomain(dbUser)
}

func (r *UserRepository) Create(
	ctx context.Context,
	u *identity.User,
) error {
	if u == nil {
		return errors.New("user is nil")
	}

	dbUser, err := r.queries.CreateUser(ctx, userssqlc.CreateUserParams{
		AuthProvider: u.AuthProvider,
		AuthSubject:  u.AuthSubject,
		Email:        u.Email,
		Role:         string(u.Role),
	})
	if err != nil {
		return err
	}

	created, err := dbUserToDomain(dbUser)
	if err != nil {
		return err
	}

	*u = *created

	return nil
}

func (r *UserRepository) UpdateAuthIdentity(
	ctx context.Context,
	id string,
	provider, subject string,
) (*identity.User, error) {
	dbUser, err := r.queries.UpdateUserAuthIdentity(ctx, userssqlc.UpdateUserAuthIdentityParams{
		AuthProvider: provider,
		AuthSubject:  subject,
		ID:           ToPgUUID(id),
	})
	if err != nil {
		return nil, err
	}

	return dbUserToDomain(dbUser)
}

func dbUserToDomain(u userssqlc.AppUser) (*identity.User, error) {
	if !u.ID.Valid {
		return nil, fmt.Errorf("user id is null")
	}

	createdAt, err := MustTime(u.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := MustTime(u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &identity.User{
		ID:           FromPgUUID(u.ID),
		AuthProvider: u.AuthProvider,
		AuthSubject:  u.AuthSubject,
		Email:        u.Email,
		Role:         identity.Role(u.Role),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}
