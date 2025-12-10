// internal/adapters/secondary/database/supabase/user_repository.go
package supabase

import (
	"context"
	"errors"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var _ repositories.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	client *Client
}

func NewUserRepository(client *Client) *UserRepository {
	return &UserRepository{client: client}
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var idUUID, subjectUUID uuid.UUID
	var roleStr string

	if err := row.Scan(
		&idUUID,
		&u.AuthProvider,
		&subjectUUID,
		&u.Email,
		&roleStr,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	u.ID = idUUID.String()
	u.AuthSubject = subjectUUID.String()
	u.Role = domain.Role(roleStr)

	return &u, nil
}

func (r *UserRepository) FindByAuthIdentity(
	ctx context.Context,
	provider, subject string,
) (*domain.User, error) {
	const query = `
		select id, 
				auth_provider, 
				auth_subject, 
				email, 
				role, 
				created_at, 
				updated_at
	from app_users
	where auth_provider = $1 and auth_subject = $2
	`
	row := r.client.Pool().QueryRow(ctx, query, provider, subject)

	u, err := scanUser(row)
	if err != nil {
		// IMPORTANTE: tratar "não encontrado" como (nil, nil)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		select id,
				auth_provider,
				auth_subject,
				email,
				role,
				created_at,
				updated_at
		from app_users
		where email = $1
	`
	row := r.client.Pool().QueryRow(ctx, query, email)

	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	const query = `
		select id, auth_provider, auth_subject, email, role, created_at, updated_at
		from app_users
		where id = $1
	`
	row := r.client.Pool().QueryRow(ctx, query, id)

	u, err := scanUser(row)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) Create(
	ctx context.Context,
	u *domain.User,
) error {
	if u == nil {
		return errors.New("user is nil")
	}

	const query = `
		insert into app_users (auth_provider, auth_subject, email, role)
		values ($1, $2, $3, $4)
		returning id, created_at, updated_at
	`

	var idUUID uuid.UUID

	err := r.client.Pool().QueryRow(ctx, query,
		u.AuthProvider,
		u.AuthSubject,
		u.Email,
		string(u.Role),
	).Scan(
		&idUUID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return err
	}

	u.ID = idUUID.String()

	return nil
}

func (r *UserRepository) UpdateAuthIdentity(
	ctx context.Context,
	id, provider, subject string,
) (*domain.User, error) {
	const query = `
		update app_users
		set auth_provider = $1,
			auth_subject = $2,
			updated_at = now()
		where id = $3
		returning id, auth_provider, auth_subject, email, role, created_at, updated_at
	`

	row := r.client.Pool().QueryRow(ctx, query, provider, subject, id)

	return scanUser(row)
}
