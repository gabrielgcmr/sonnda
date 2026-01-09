// internal/app/services/patient/service_impl.go
package patientsvc

import (
	"context"

	"sonnda-api/internal/app/services/authorization"
	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/rbac"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/ports/repository"

	"github.com/google/uuid"
)

type service struct {
	repo repository.Patient
	auth authorization.Authorizer
}

var _ Service = (*service)(nil)

func New(repo repository.Patient, auth authorization.Authorizer) Service {
	return &service{
		repo: repo,
		auth: auth,
	}
}

func (s *service) Create(ctx context.Context, currentUser *user.User, input CreateInput) (*patient.Patient, error) {
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

func (s *service) Get(ctx context.Context, currentUser *user.User, id uuid.UUID) (*patient.Patient, error) {

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoError("patientRepo.FindByID", err)
	}
	if p == nil {
		return nil, patientNotFound()
	}
	return p, nil
}

func (s *service) Update(ctx context.Context, currentUser *user.User, id uuid.UUID, input UpdateInput) (*patient.Patient, error) {

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

func (s *service) SoftDelete(ctx context.Context, currentUser *user.User, id uuid.UUID) error {

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

func (s *service) HardDelete(ctx context.Context, currentUser *user.User, id uuid.UUID) error {
	if err := s.auth.Require(ctx, currentUser, rbac.ActionHardDeletePatient, &id); err != nil {
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

func (s *service) List(ctx context.Context, currentUser *user.User, limit, offset int) ([]*patient.Patient, error) {
	if err := s.auth.Require(ctx, currentUser, rbac.ActionListPatients, nil); err != nil {
		return nil, err
	}

	rows, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, mapRepoError("patientRepo.List", err)
	}

	out := make([]*patient.Patient, len(rows))
	for i := range rows {
		out[i] = &rows[i]
	}

	return out, nil
}
