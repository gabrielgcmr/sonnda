package usersvc

import (
	"context"
	"errors"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/ports/repository"

	"github.com/google/uuid"
)

type service struct {
	userRepo repository.User
}

var _ Service = (*service)(nil)

func New(userRepo repository.User) Service {
	return &service{userRepo: userRepo}
}

func (s *service) Create(ctx context.Context, input UserCreateInput) (*user.User, error) {
	newUser, err := user.NewUser(user.NewUserParams{
		AuthProvider: input.Provider,
		AuthSubject:  input.Subject,
		Email:        input.Email,
		AccountType:  input.AccountType,
		FullName:     input.FullName,
		BirthDate:    input.BirthDate,
		CPF:          input.CPF,
		Phone:        input.Phone,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, mapRepoError("userRepo.Create", err)
	}

	return newUser, nil
}

func (s *service) Update(ctx context.Context, input UserUpdateInput) (*user.User, error) {
	existingUser, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, mapRepoError("userRepo.FindByID", err)
	}
	if existingUser == nil {
		return nil, userNotFound()
	}

	changed, err := existingUser.ApplyUpdate(user.UpdateUserParams{
		FullName:  input.FullName,
		BirthDate: input.BirthDate,
		CPF:       input.CPF,
		Phone:     input.Phone,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	if !changed {
		return existingUser, nil
	}

	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, mapRepoError("userRepo.Update", err)
	}

	return existingUser, nil
}

func (s *service) Delete(ctx context.Context, userID uuid.UUID) error {
	existing, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return mapRepoError("userRepo.FindByID", err)
	}
	if existing == nil {
		return userNotFound()
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return mapRepoError("userRepo.Delete", err)
	}

	return nil
}

func (s *service) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	// NOTE: SoftDelete is intentionally idempotent.
	//
	// We first load the user to return a proper NOT_FOUND when it truly doesn't exist.
	// Then we execute the delete; if the repository reports NOT_FOUND at this stage
	// (e.g. already deleted or a race where another request deleted it), we treat it as success.
	existing, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return mapRepoError("userRepo.FindByID", err)
	}
	if existing == nil {
		return userNotFound()
	}

	if err := s.userRepo.SoftDelete(ctx, userID); err != nil {
		mapped := mapRepoError("userRepo.SoftDelete", err)
		var appErr *apperr.AppError
		if errors.As(mapped, &appErr) && appErr != nil && appErr.Code == apperr.NOT_FOUND {
			return nil
		}
		return mapped
	}

	return nil
}
