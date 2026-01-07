package patient

import (
	"errors"
	"fmt"
	"sonnda-api/internal/adapters/outbound/persistence/repository/helpers"
)

var (
	ErrPatientAlreadyExists = errors.New("patient already exists")
	ErrRepositoryFailure    = errors.New("repository failure")
	ErrNotFound             = errors.New("not found")
)

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	//Violação de unicidade
	if helpers.IsUniqueViolationError(err) {
		return ErrPatientAlreadyExists
	}

	return fmt.Errorf("%w: %v", ErrRepositoryFailure, err)
}
