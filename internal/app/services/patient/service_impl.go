// internal/app/services/patient/service_impl.go
package patientsvc

import (
	"context"

	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

type service struct {
	repo   repositories.PatientRepository
	policy AccessPolicy
}

var _ Service = (*service)(nil)

func New(repo repositories.PatientRepository, policy AccessPolicy) Service {
	if policy == nil {
		policy = AllowAllPolicy{}
	}
	return &service{repo: repo, policy: policy}
}

func (s *service) Create(ctx context.Context, currentUser *user.User, input CreateInput) (*patient.Patient, error) {
	if err := s.policy.CanCreate(ctx, currentUser, input); err != nil {
		return nil, err
	}

	p, err := patient.NewPatient(patient.NewPatientParams{
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
		return nil, err
	}

	// MVP: checagem explícita (honesta) — alternativa é confiar no UNIQUE e mapear erro.
	existing, err := s.repo.FindByCPF(ctx, p.CPF)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, patient.ErrCPFAlreadyExists
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) GetByID(ctx context.Context, currentUser *user.User, id uuid.UUID) (*patient.Patient, error) {
	if err := s.policy.CanRead(ctx, currentUser, id); err != nil {
		return nil, err
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
	}
	return p, nil
}

func (s *service) UpdateByID(ctx context.Context, currentUser *user.User, id uuid.UUID, input UpdateInput) (*patient.Patient, error) {
	if err := s.policy.CanUpdate(ctx, currentUser, id); err != nil {
		return nil, err
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, patient.ErrPatientNotFound
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
		return nil, err
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) SoftDeleteByID(ctx context.Context, currentUser *user.User, id uuid.UUID) error {
	if err := s.policy.CanDelete(ctx, currentUser, id); err != nil {
		return err
	}

	// opcional: checar existência para erro melhor
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return patient.ErrPatientNotFound
	}

	return s.repo.SoftDelete(ctx, id)
}

func (s *service) List(ctx context.Context, currentUser *user.User, limit, offset int) ([]*patient.Patient, error) {
	if err := s.policy.CanList(ctx, currentUser); err != nil {
		return nil, err
	}

	rows, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	patients := make([]*patient.Patient, 0, len(rows))
	for i := range rows {
		patients = append(patients, &rows[i])
	}

	return patients, nil

}

// ListByName implements [Service].
func (s *service) ListByName(ctx context.Context, currentUser *user.User, name string, limit int, offset int) ([]*patient.Patient, error) {
	panic("unimplemented")
}
