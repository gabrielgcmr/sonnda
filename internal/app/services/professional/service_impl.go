package professionalsvc

import (
	"context"
	"errors"

	"sonnda-api/internal/app/interfaces/repositories"
	"sonnda-api/internal/domain/model/user/professional"

	"github.com/google/uuid"
)

type service struct {
	repo repositories.ProfessionalRepository
}

var _ Service = (*service)(nil)

func New(repo repositories.ProfessionalRepository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, input CreateInput) (*professional.Professional, error) {
	if s == nil || s.repo == nil {
		return nil, mapDomainError(errors.New("professional repository not configured"))
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
		return nil, mapInfraError("profRepo.Create", err)
	}

	return prof, nil
}

func (s *service) GetByID(ctx context.Context, profileID uuid.UUID) (*professional.Professional, error) {
	if s == nil || s.repo == nil {
		return nil, mapDomainError(errors.New("professional repository not configured"))
	}

	p, err := s.repo.FindByID(ctx, profileID)
	if err != nil {
		return nil, mapInfraError("profRepo.FindByID", err)
	}
	if p == nil {
		return nil, mapDomainError(professional.ErrProfileNotFound)
	}
	return p, nil
}

func (s *service) GetByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error) {
	if s == nil || s.repo == nil {
		return nil, mapDomainError(errors.New("professional repository not configured"))
	}

	p, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, mapInfraError("profRepo.FindByUserID", err)
	}
	if p == nil {
		return nil, mapDomainError(professional.ErrProfileNotFound)
	}
	return p, nil
}

func (s *service) DeleteByID(ctx context.Context, profileID uuid.UUID) error {
	if s == nil || s.repo == nil {
		return mapDomainError(errors.New("professional repository not configured"))
	}

	existing, err := s.repo.FindByID(ctx, profileID)
	if err != nil {
		return mapInfraError("profRepo.FindByID", err)
	}
	if existing == nil {
		return mapDomainError(professional.ErrProfileNotFound)
	}

	if err := s.repo.Delete(ctx, profileID); err != nil {
		return mapInfraError("profRepo.Delete", err)
	}
	return nil
}
