// internal/infrastructure/persistence/supabase/user_repository.go
package user

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sonnda-api/internal/app/interfaces/repositories"
	"sonnda-api/internal/domain/model/user"
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

// Create implements [repositories.UserRepository].
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	params := usersqlc.CreateUserParams{
		ID:           u.ID,
		AuthProvider: u.AuthProvider,
		AuthSubject:  u.AuthSubject,
		Email:        u.Email,
		FullName:     u.FullName,
		BirthDate:    helpers.FromRequiredDateToPgDate(u.BirthDate),
		Cpf:          u.CPF,
		Phone:        u.Phone,
		AccountType:  string(u.AccountType),
		CreatedAt:    helpers.FromRequiredTimestamptzToPgTimestamptz(u.CreatedAt),
		UpdatedAt:    helpers.FromRequiredTimestamptzToPgTimestamptz(u.UpdatedAt),
	}

	if err := r.queries.CreateUser(ctx, params); err != nil {
		return mapRepositoryError(err)
	}

	// Não sobrescreve a entidade; app-source-of-truth mantém valores do domínio.
	return nil
}

// Delete implements [repositories.UserRepository].
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return mapRepositoryError(err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return mapRepositoryError(err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// FindByAuthIdentity implements [repositories.UserRepository].
func (r *UserRepository) FindByAuthIdentity(ctx context.Context, provider string, subject string) (*user.User, error) {
	row, err := r.queries.FindUserByAuthIdentity(ctx, usersqlc.FindUserByAuthIdentityParams{
		AuthProvider: provider,
		AuthSubject:  subject,
	})
	if err != nil {
		return nil, mapRepositoryError(err)
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
		AccountType:  user.AccountType(row.AccountType),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

// FindByCPF implements [repositories.UserRepository].
func (r *UserRepository) FindByCPF(ctx context.Context, cpf string) (*user.User, error) {
	row, err := r.queries.FindUserByCPF(ctx, cpf)
	if err != nil {
		return nil, mapRepositoryError(err)
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
		AccountType:  user.AccountType(row.AccountType),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

// FindByID implements [repositories.UserRepository].
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row, err := r.queries.FindUserByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError(err)
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
		AccountType:  user.AccountType(row.AccountType),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

// Update implements [repositories.UserRepository].
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user is nil")
	}

	birthDate := u.BirthDate
	row, err := r.queries.UpdateUser(ctx, usersqlc.UpdateUserParams{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		BirthDate: helpers.FromNullableDateToPgDate(&birthDate),
		Cpf:       u.CPF,
		Phone:     u.Phone,
		UpdatedAt: helpers.FromRequiredTimestamptzToPgTimestamptz(u.UpdatedAt),
	})
	if err != nil {
		return mapRepositoryError(err)
	}

	u.ID = row.ID
	u.AuthProvider = row.AuthProvider
	u.AuthSubject = row.AuthSubject
	u.Email = row.Email
	u.FullName = row.FullName
	u.BirthDate = row.BirthDate.Time
	u.CPF = row.Cpf
	u.Phone = row.Phone
	u.AccountType = user.AccountType(row.AccountType)
	u.CreatedAt = row.CreatedAt.Time
	u.UpdatedAt = row.UpdatedAt.Time

	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row, err := r.queries.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, mapRepositoryError(err)
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
		AccountType:  user.AccountType(row.AccountType),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}
