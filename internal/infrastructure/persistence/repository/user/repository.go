// internal/infrastructure/persistence/supabase/user_repository.go
package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"sonnda-api/internal/domain/entities/rbac"
	"sonnda-api/internal/domain/entities/user"
	"sonnda-api/internal/domain/ports/repositories"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
	"sonnda-api/internal/infrastructure/persistence/repository/helpers"
	usersqlc "sonnda-api/internal/infrastructure/persistence/sqlc/generated/user"
)

var _ repositories.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	client  *db.Client
	queries *usersqlc.Queries
}

func New(client *db.Client) *UserRepository {
	return &UserRepository{
		client:  client,
		queries: usersqlc.New(client.Pool()),
	}
}

// Delete implements [repositories.UserRepository].
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.SoftDeleteUser(ctx, id)
	return err
}

// FindByAuthIdentity implements [repositories.UserRepository].
func (r *UserRepository) FindByAuthIdentity(ctx context.Context, provider string, subject string) (*user.User, error) {
	row, err := r.queries.FindUserByAuthIdentity(ctx, usersqlc.FindUserByAuthIdentityParams{
		AuthProvider: provider,
		AuthSubject:  subject,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user.User{
		ID:           row.ID,
		AuthProvider: row.AuthProvider,
		AuthSubject:  row.AuthSubject,
		Email:        row.Email,
		FullName:     row.FullName,
		BirthDate:    row.BirthDate.Time,
		CPF:          row.Cpf,
		Phone:        row.Phone,
		Role:         rbac.Role(row.Role),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

// FindByCPF implements [repositories.UserRepository].
func (r *UserRepository) FindByCPF(ctx context.Context, cpf string) (*user.User, error) {
	row, err := r.queries.FindUserByCPF(ctx, cpf)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user.User{
		ID:           row.ID,
		AuthProvider: row.AuthProvider,
		AuthSubject:  row.AuthSubject,
		Email:        row.Email,
		FullName:     row.FullName,
		BirthDate:    row.BirthDate.Time,
		CPF:          row.Cpf,
		Phone:        row.Phone,
		Role:         rbac.Role(row.Role),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

// FindByID implements [repositories.UserRepository].
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	panic("unimplemented")
}

// Save implements [repositories.UserRepository].
func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user is nil")
	}

	if u.ID == uuid.Nil {
		u.ID = uuid.Must(uuid.NewV7())
	}

	birthDate := u.BirthDate
	row, err := r.queries.CreateUser(ctx, usersqlc.CreateUserParams{
		ID:           u.ID,
		AuthProvider: u.AuthProvider,
		AuthSubject:  u.AuthSubject,
		Email:        u.Email,
		FullName:     u.FullName,
		BirthDate:    helpers.FromNullableDateToPgDate(&birthDate),
		Cpf:          u.CPF,
		Phone:        u.Phone,
		Role:         string(u.Role),
	})
	if err != nil {
		return err
	}

	u.ID = row.ID
	u.AuthProvider = row.AuthProvider
	u.AuthSubject = row.AuthSubject
	u.Email = row.Email
	u.FullName = row.FullName
	u.BirthDate = row.BirthDate.Time
	u.CPF = row.Cpf
	u.Phone = row.Phone
	u.Role = rbac.Role(row.Role)
	u.CreatedAt = row.CreatedAt.Time
	u.UpdatedAt = row.UpdatedAt.Time

	return nil
}

// Update implements [repositories.UserRepository].
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	panic("unimplemented")
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row, err := r.queries.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user.User{
		ID:           row.ID,
		AuthProvider: row.AuthProvider,
		AuthSubject:  row.AuthSubject,
		Email:        row.Email,
		FullName:     row.FullName,
		BirthDate:    row.BirthDate.Time,
		CPF:          row.Cpf,
		Phone:        row.Phone,
		Role:         rbac.Role(row.Role),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}
