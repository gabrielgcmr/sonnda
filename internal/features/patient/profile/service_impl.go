// internal/features/patient/profile/service_impl.go
package patientprofile

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patientaccess"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"

	"github.com/google/uuid"
)

type service struct {
	repo       Repository
	accessRepo repository.PatientAccessRepo
	auth       Authorizer
}

var _ Service = (*service)(nil)

func New(
	repo Repository,
	accessRepo repository.PatientAccessRepo,
	auth Authorizer,
) Service {
	return &service{
		repo:       repo,
		accessRepo: accessRepo,
		auth:       auth,
	}
}

func (s *service) Create(ctx context.Context, currentUser *accountdomain.User, input CreateInput) (*profiledomain.Patient, error) {
	if currentUser == nil {
		return nil, apperr.Unauthorized("autenticação necessária")
	}
	if s.accessRepo == nil {
		return nil, apperr.Internal("erro inesperado", errors.New("patient access repository not configured"))
	}

	relationType, err := relationTypeForCreator(currentUser)
	if err != nil {
		return nil, err
	}
	if input.RelationType != nil {
		relationType = *input.RelationType
	}

	newPatient, err := profiledomain.NewPatient(profiledomain.NewPatientParams{
		UserID:    input.UserID,
		CPF:       input.CPF,
		CNS:       input.CNS,
		FullName:  input.FullName,
		BirthDate: input.BirthDate,
		Gender:    input.Gender,
		Race:      input.Race,
		Phone:     input.Phone,
		AvatarURL: input.AvatarURL,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	access, err := patientaccess.NewPatientAccess(
		newPatient.ID,
		currentUser.ID,
		relationType,
		&currentUser.ID,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, apperr.Internal("erro inesperado", err)
	}
	if err := s.repo.CreateWithAccess(ctx, newPatient, access); err != nil {
		return nil, mapRepoError("patientRepo.CreateWithAccess", err)
	}

	return newPatient, nil
}

func (s *service) Get(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) (*profiledomain.Patient, error) {
	if err := s.auth.RequirePatientAccess(ctx, currentUser, id); err != nil {
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
	if err := s.auth.RequirePatientAccess(ctx, currentUser, id); err != nil {
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
	if err := s.auth.RequirePatientAccess(ctx, currentUser, id); err != nil {
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
	if err := s.auth.RequirePatientAccess(ctx, currentUser, id); err != nil {
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

func relationTypeForCreator(currentUser *accountdomain.User) (patientaccess.RelationshipType, error) {
	switch currentUser.AccountType.Normalize() {
	case accountdomain.AccountTypeProfessional:
		return patientaccess.RelationshipTypeProfessional, nil
	case accountdomain.AccountTypeBasicCare:
		return patientaccess.RelationshipTypeCaregiver, nil
	default:
		return "", apperr.Internal("erro inesperado", fmt.Errorf("unsupported account type: %s", currentUser.AccountType))
	}
}
