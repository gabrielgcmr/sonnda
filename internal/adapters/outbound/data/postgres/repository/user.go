// internal/adapters/outbound/data/postgres/repository/user.go
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/gabrielgcmr/sonnda/internal/adapters/outbound/data/postgres/repository/db"
	usersqlc "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/data/postgres/sqlc/generated/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
)

var _ ports.UserRepo = (*UserRepository)(nil)

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

// Create implements [ports.User].
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	params := usersqlc.CreateUserParams{
		ID:           u.ID,
		AuthProvider: u.AuthProvider,
		AuthSubject:  u.AuthSubject,
		Email:        u.Email,
		FullName:     u.FullName,
		BirthDate:    FromRequiredDateToPgDate(u.BirthDate),
		Cpf:          u.CPF,
		Phone:        u.Phone,
		AccountType:  string(u.AccountType),
		CreatedAt:    FromRequiredTimestamptzToPgTimestamptz(u.CreatedAt),
		UpdatedAt:    FromRequiredTimestamptzToPgTimestamptz(u.UpdatedAt),
	}

	if err := r.queries.CreateUser(ctx, params); err != nil {
		if IsUniqueViolationError(err) {
			return ErrUserAlreadyExists
		}
		return errors.Join(ErrRepositoryFailure, err)
	}

	// NÇœo sobrescreve a entidade; app-source-of-truth mantÇ¸m valores do domÇðnio.
	return nil
}

// Delete implements [ports.User].
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return errors.Join(ErrRepositoryFailure, err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return errors.Join(ErrRepositoryFailure, err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// FindByAuthIdentity implements [ports.User].
func (r *UserRepository) FindByAuthIdentity(ctx context.Context, provider string, subject string) (*user.User, error) {
	row, err := r.queries.FindUserByAuthIdentity(ctx, usersqlc.FindUserByAuthIdentityParams{
		AuthProvider: provider,
		AuthSubject:  subject,
	})
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, errors.Join(ErrRepositoryFailure, err)
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

// FindByCPF implements [ports.User].
func (r *UserRepository) FindByCPF(ctx context.Context, cpf string) (*user.User, error) {
	row, err := r.queries.FindUserByCPF(ctx, cpf)
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, errors.Join(ErrRepositoryFailure, err)
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

// FindByID implements [ports.User].
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row, err := r.queries.FindUserByID(ctx, id)
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, errors.Join(ErrRepositoryFailure, err)
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

// Update implements [ports.User].
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	birthDate := u.BirthDate
	row, err := r.queries.UpdateUser(ctx, usersqlc.UpdateUserParams{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		BirthDate: FromNullableDateToPgDate(&birthDate),
		Cpf:       u.CPF,
		Phone:     u.Phone,
		UpdatedAt: FromRequiredTimestamptzToPgTimestamptz(u.UpdatedAt),
	})
	if err != nil {
		if IsPgNotFound(err) {
			return ErrUserNotFound
		}
		if IsUniqueViolationError(err) {
			return ErrUserAlreadyExists
		}
		return errors.Join(ErrRepositoryFailure, err)
	}

	u.ID = row.ID
	u.Email = row.Email
	u.FullName = row.FullName
	u.BirthDate = row.BirthDate.Time
	u.CPF = row.Cpf
	u.Phone = row.Phone
	u.UpdatedAt = row.UpdatedAt.Time

	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row, err := r.queries.FindUserByEmail(ctx, email)
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, errors.Join(ErrRepositoryFailure, err)
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
