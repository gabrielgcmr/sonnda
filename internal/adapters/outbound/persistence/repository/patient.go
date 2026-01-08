package repository

import (
	"context"
	"errors"
	"time"

	"sonnda-api/internal/adapters/outbound/persistence/repository/db"

	patientsqlc "sonnda-api/internal/adapters/outbound/persistence/sqlc/generated/patient"
	"sonnda-api/internal/domain/model/demographics"
	"sonnda-api/internal/domain/model/patient"
	"sonnda-api/internal/domain/ports/repository"

	"github.com/google/uuid"
)

type PatientRepository struct {
	client  *db.Client
	queries *patientsqlc.Queries
}

var _ repository.Patient = (*PatientRepository)(nil)

func NewPatientRepository(client *db.Client) repository.Patient {
	return &PatientRepository{
		client:  client,
		queries: patientsqlc.New(client.Pool()),
	}
}

// Create implements [repository.Patient].
func (r *PatientRepository) Create(ctx context.Context, p *patient.Patient) error {
	params := patientsqlc.CreatePatientParams{
		ID:          p.ID,
		OwnerUserID: FromNullableUUIDToPgUUID(p.OwnerUserID),
		Cpf:         p.CPF,
		Cns:         FromNullableStringToPgText(p.CNS),
		FullName:    p.FullName,
		BirthDate:   FromRequiredDateToPgDate(p.BirthDate),
		Gender:      string(p.Gender),
		Race:        string(p.Race),
		Phone:       FromNullableStringToPgText(p.Phone),
		AvatarUrl:   FromNullableStringToPgText(&p.AvatarURL),
	}

	row, err := r.queries.CreatePatient(ctx, params)
	if err != nil {
		if IsUniqueViolationError(err) {
			return ErrPatientAlreadyExists
		}
		return errors.Join(ErrRepositoryFailure, err)
	}

	p.ID = row.ID
	p.OwnerUserID = FromPgUUIDToNullableUUID(row.OwnerUserID)
	p.CPF = row.Cpf
	p.CNS = FromPgTextToNullableString(row.Cns)
	p.FullName = row.FullName
	p.BirthDate = row.BirthDate.Time
	p.Gender = demographics.Gender(row.Gender)
	p.Race = demographics.Race(row.Race)
	p.AvatarURL = row.AvatarUrl.String
	p.Phone = FromPgTextToNullableString(row.Phone)
	p.CreatedAt = row.CreatedAt.Time
	p.UpdatedAt = row.UpdatedAt.Time

	return nil
}

// SoftDelete implements [repository.Patient].
func (p *PatientRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FindByCPF implements [repository.Patient].
func (p *PatientRepository) FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	row, err := p.queries.GetPatientByCPF(ctx, cpf)
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, errors.Join(ErrRepositoryFailure, err)
	}

	return &patient.Patient{
		ID:          row.ID,
		OwnerUserID: FromPgUUIDToNullableUUID(row.OwnerUserID),
		CPF:         row.Cpf,
		CNS:         FromPgTextToNullableString(row.Cns),
		FullName:    row.FullName,
		BirthDate:   row.BirthDate.Time,
		Gender:      demographics.Gender(row.Gender),
		Race:        demographics.Race(row.Race),
		AvatarURL:   row.AvatarUrl.String,
		Phone:       FromPgTextToNullableString(row.Phone),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

// FindByID implements [repository.Patient].
func (p *PatientRepository) FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error) {
	row, err := p.queries.GetPatientByID(ctx, id)
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, errors.Join(ErrRepositoryFailure, err)
	}

	return &patient.Patient{
		ID:          row.ID,
		OwnerUserID: FromPgUUIDToNullableUUID(row.OwnerUserID),
		CPF:         row.Cpf,
		CNS:         FromPgTextToNullableString(row.Cns),
		FullName:    row.FullName,
		BirthDate:   row.BirthDate.Time,
		Gender:      demographics.Gender(row.Gender),
		Race:        demographics.Race(row.Race),
		AvatarURL:   row.AvatarUrl.String,
		Phone:       FromPgTextToNullableString(row.Phone),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

// List implements [repository.Patient].
func (p *PatientRepository) List(ctx context.Context, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// ListByBirthDate implements [repository.Patient].
func (p *PatientRepository) ListByBirthDate(ctx context.Context, birthDate time.Time, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// ListByIDs implements [repository.Patient].
func (p *PatientRepository) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]patient.Patient, error) {
	panic("unimplemented")
}

// ListByName implements [repository.Patient].
func (p *PatientRepository) ListByName(ctx context.Context, name string, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// Update implements [repository.Patient].
func (p *PatientRepository) Update(ctx context.Context, patient *patient.Patient) error {
	panic("unimplemented")
}
