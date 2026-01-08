// internal/app/services/patient/service_impl.go
package patientsvc

import (
	"context"
	"errors"

	"sonnda-api/internal/app/apperr"
	"sonnda-api/internal/app/services/authorization"
	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/rbac"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/ports/repository"

	"github.com/google/uuid"
)

type coreService struct {
	repo repository.PatientRepository
}

func newCore(repo repository.PatientRepository) *coreService {
	return &coreService{repo: repo}
}

func (s *coreService) Create(ctx context.Context, input CreateInput) (*patient.Patient, error) {
	newPatient, err := patient.NewPatient(patient.NewPatientParams{
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

	if err := s.repo.Create(ctx, newPatient); err != nil {
		return nil, mapRepoError("patientRepo.Create", err)
	}

	return newPatient, nil
}

func (s *coreService) GetByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoError("patientRepo.FindByID", err)
	}
	if p == nil {
		return nil, patientNotFound(nil)
	}
	return p, nil
}

func (s *coreService) UpdateByID(ctx context.Context, id uuid.UUID, input UpdateInput) (*patient.Patient, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoError("patientRepo.FindByID", err)
	}
	if p == nil {
		return nil, patientNotFound(nil)
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

func (s *coreService) SoftDeleteByID(ctx context.Context, id uuid.UUID) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return mapRepoError("patientRepo.FindByID", err)
	}
	if p == nil {
		return patientNotFound(nil)
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return mapRepoError("patientRepo.SoftDelete", err)
	}
	return nil
}

func (s *coreService) List(ctx context.Context, limit, offset int) ([]*patient.Patient, error) {
	rows, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, mapRepoError("patientRepo.List", err)
	}

	patients := make([]*patient.Patient, 0, len(rows))
	for i := range rows {
		patients = append(patients, &rows[i])
	}
	return patients, nil
}

func (s *coreService) ListByName(ctx context.Context, name string, limit, offset int) ([]*patient.Patient, error) {
	rows, err := s.repo.ListByName(ctx, name, limit, offset)
	if err != nil {
		return nil, mapRepoError("patientRepo.ListByName", err)
	}

	patients := make([]*patient.Patient, 0, len(rows))
	for i := range rows {
		patients = append(patients, &rows[i])
	}
	return patients, nil
}

type service struct {
	core *coreService
	auth authorization.Authorizer
}

var _ Service = (*service)(nil)

func New(repo repository.PatientRepository, auth authorization.Authorizer) Service {
	return &service{
		core: newCore(repo),
		auth: auth,
	}
}

func (s *service) Create(ctx context.Context, currentUser *user.User, input CreateInput) (*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionCreatePatient, nil); err != nil {
		return nil, err
	}

	return s.core.Create(ctx, input)
}

func (s *service) GetByID(ctx context.Context, currentUser *user.User, id uuid.UUID) (*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionReadPatient, &id); err != nil {
		return nil, err
	}

	p, err := s.core.GetByID(ctx, id)
	if err != nil {
		return nil, maskNotFoundAsForbidden(err)
	}
	return p, nil
}

func (s *service) UpdateByID(ctx context.Context, currentUser *user.User, id uuid.UUID, input UpdateInput) (*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionUpdatePatient, &id); err != nil {
		return nil, err
	}

	p, err := s.core.UpdateByID(ctx, id, input)
	if err != nil {
		return nil, maskNotFoundAsForbidden(err)
	}
	return p, nil
}

func (s *service) SoftDeleteByID(ctx context.Context, currentUser *user.User, id uuid.UUID) error {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionSoftDeletePatient, &id); err != nil {
		return err
	}

	if err := s.core.SoftDeleteByID(ctx, id); err != nil {
		return maskNotFoundAsForbidden(err)
	}
	return nil
}

func (s *service) HardDeleteByID(ctx context.Context, currentUser *user.User, id uuid.UUID) error {
	panic("unimplemented")
}

func (s *service) List(ctx context.Context, currentUser *user.User, limit, offset int) ([]*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionListPatients, nil); err != nil {
		return nil, err
	}

	return s.core.List(ctx, limit, offset)
}

func (s *service) ListByName(ctx context.Context, currentUser *user.User, name string, limit, offset int) ([]*patient.Patient, error) {
	if s == nil || s.core == nil || s.core.repo == nil || s.auth == nil {
		return nil, apperr.Internal("serviço indisponível", errors.New("patient service not configured"))
	}

	if err := s.auth.Require(ctx, currentUser, rbac.ActionListPatients, nil); err != nil {
		return nil, err
	}

	return s.core.ListByName(ctx, name, limit, offset)
}

func maskNotFoundAsForbidden(err error) error {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil && appErr.Code == apperr.NOT_FOUND {
		return apperr.Forbidden("acesso negado")
	}
	return err
}
