package patientsvc

import (
	"context"

	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/ports/repository"

	"github.com/google/uuid"
)

type coreService struct {
	repo repository.Patient
}

func newCore(repo repository.Patient) *coreService {
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
