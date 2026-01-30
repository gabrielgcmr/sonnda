// internal/application/services/professional/service_impl.go
package professionalsvc

import (
	"context"
	"errors"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/professional"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/google/uuid"
)

type service struct {
	repo repository.ProfessionalRepo
}

var _ Service = (*service)(nil)

func New(repo repository.ProfessionalRepo) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, input CreateInput) (*professional.Professional, error) {
	if s == nil || s.repo == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("professional repository not configured"))
	}

	prof, err := professional.NewProfessional(professional.NewProfessionalParams{
		UserID:             input.UserID,
		Kind:               input.Kind,
		RegistrationNumber: input.RegistrationNumber,
		RegistrationIssuer: input.RegistrationIssuer,
		RegistrationState:  input.RegistrationState,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	if err := s.repo.Create(ctx, prof); err != nil {
		return nil, mapRepoError("profRepo.Create", err)
	}

	return prof, nil
}

func (s *service) GetByID(ctx context.Context, profileID uuid.UUID) (*professional.Professional, error) {
	if s == nil || s.repo == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("professional repository not configured"))
	}

	p, err := s.repo.FindByID(ctx, profileID)
	if err != nil {
		return nil, mapRepoError("profRepo.FindByID", err)
	}
	if p == nil {
		return nil, professionalNotFound(nil)
	}
	return p, nil
}

func (s *service) GetByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error) {
	if s == nil || s.repo == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("professional repository not configured"))
	}

	p, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, mapRepoError("profRepo.FindByUserID", err)
	}
	if p == nil {
		return nil, professionalNotFound(nil)
	}
	return p, nil
}

func (s *service) Delete(ctx context.Context, profileID uuid.UUID) error {
	if s == nil || s.repo == nil {
		return apperr.Internal("serviço indisponível", errors.New("professional repository not configured"))
	}

	existing, err := s.repo.FindByID(ctx, profileID)
	if err != nil {
		return mapRepoError("profRepo.FindByID", err)
	}
	if existing == nil {
		return professionalNotFound(nil)
	}

	if err := s.repo.Delete(ctx, profileID); err != nil {
		return mapRepoError("profRepo.Delete", err)
	}
	return nil
}
