// internal/features/patient/profile/service_impl.go
package patientprofile

import (
	"context"
	"errors"
	"fmt"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/google/uuid"
)

type service struct {
	repo          Repository
	accessRepo    patientaccess.Repository
	accessChecker AccessChecker
}

var _ Service = (*service)(nil)

func New(
	repo Repository,
	accessRepo patientaccess.Repository,
	accessChecker AccessChecker,
) Service {
	return &service{
		repo:          repo,
		accessRepo:    accessRepo,
		accessChecker: accessChecker,
	}
}

func (s *service) Get(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) (*profiledomain.Patient, error) {
	if err := s.accessChecker.RequireAccess(ctx, currentAccountID(currentUser), id); err != nil {
		return nil, err
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoError("patientRepo.FindByID", err)
	}
	if p == nil {
		return nil, patientNotFound()
	}
	return p, nil
}

func (s *service) Update(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID, input UpdateInput) (*profiledomain.Patient, error) {
	if err := s.accessChecker.RequireAccess(ctx, currentAccountID(currentUser), id); err != nil {
		return nil, err
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoError("patientRepo.FindByID", err)
	}
	if p == nil {
		return nil, patientNotFound()
	}

	p.ApplyUpdate(
		input.FullName,
		input.Phone,
		input.AvatarURL,
		input.Gender,
		input.Race,
		input.CNS,
	)

	if err := p.Validate(); err != nil {
		return nil, mapDomainError(err)
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, mapRepoError("patientRepo.Update", err)
	}
	return p, nil
}

func (s *service) SoftDelete(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) error {
	if err := s.accessChecker.RequireAccess(ctx, currentAccountID(currentUser), id); err != nil {
		return err
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return mapRepoError("patientRepo.FindByID", err)
	}
	if p == nil {
		return patientNotFound()
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return mapRepoError("patientRepo.SoftDelete", err)
	}
	return nil
}

func (s *service) HardDelete(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) error {
	if err := s.accessChecker.RequireAccess(ctx, currentAccountID(currentUser), id); err != nil {
		return err
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return mapRepoError("patientRepo.FindByID", err)
	}
	if p == nil {
		return patientNotFound()
	}

	if err := s.repo.HardDelete(ctx, id); err != nil {
		return mapRepoError("patientRepo.HardDelete", err)
	}
	return nil
}

func currentAccountID(currentUser *accountdomain.User) uuid.UUID {
	if currentUser == nil {
		return uuid.Nil
	}
	return currentUser.ID
}

func (s *service) ListMyPatients(ctx context.Context, currentUser *accountdomain.User, limit, offset int) ([]*profiledomain.Patient, error) {
	if currentUser == nil {
		return nil, apperr.Unauthorized("autenticação necessária")
	}

	if s.accessRepo == nil {
		return nil, apperr.Internal("erro inesperado", errors.New("patient access repository not configured"))
	}

	accessible, _, err := s.accessRepo.ListAccessiblePatientsByUser(ctx, currentUser.ID, limit, offset)
	if err != nil {
		return nil, &apperr.AppError{
			Kind:    apperr.INFRA_DATABASE_ERROR,
			Message: "falha técnica",
			Cause:   fmt.Errorf("patientAccessRepo.ListAccessiblePatientsByUser: %w", err),
		}
	}

	out := make([]*profiledomain.Patient, 0, len(accessible))
	for _, row := range accessible {
		p, err := s.repo.FindByID(ctx, row.PatientID)
		if err != nil {
			return nil, mapRepoError("patientRepo.FindByID", err)
		}
		if p == nil {
			continue
		}
		out = append(out, p)
	}

	return out, nil
}
