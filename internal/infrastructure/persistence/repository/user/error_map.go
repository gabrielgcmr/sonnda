package user

import (
	"errors"
	"fmt"
	"sonnda-api/internal/infrastructure/persistence/repository/helpers"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrRepositoryFailure = errors.New("repository failure")
	ErrNotFound          = errors.New("not found")
)

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	//Violação de unicidade
	if helpers.IsUniqueViolationError(err) {
		return ErrUserAlreadyExists
	}

	return fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
}
