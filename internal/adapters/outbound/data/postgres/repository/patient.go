// internal/adapters/outbound/data/postgres/repository/patient.go
package repository

import (
	"context"
	"errors"

	"github.com/gabrielgcmr/sonnda/internal/adapters/outbound/data/postgres/repository/db"

	patientsqlc "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/data/postgres/sqlc/generated/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/demographics"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"

	"github.com/google/uuid"
)

type PatientRepository struct {
	client  *db.Client
	queries *patientsqlc.Queries
}

// FindByName implements [ports.Patient].
func (r *PatientRepository) FindByName(ctx context.Context, name string) ([]patient.Patient, error) {
	panic("unimplemented")
}

// SearchByName implements [ports.Patient].
func (r *PatientRepository) SearchByName(ctx context.Context, name string, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// HardDelete implements [ports.Patient].
func (r *PatientRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// List implements [ports.Patient].
func (r *PatientRepository) List(ctx context.Context, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// Update implements [ports.Patient].
func (r *PatientRepository) Update(ctx context.Context, patient *patient.Patient) error {
	panic("unimplemented")
}

var _ ports.PatientRepo = (*PatientRepository)(nil)

func NewPatientRepository(client *db.Client) ports.PatientRepo {
	return &PatientRepository{
		client:  client,
		queries: patientsqlc.New(client.Pool()),
	}
}

// Create implements [ports.Patient].
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

// SoftDelete implements [ports.Patient].
func (p *PatientRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FindByCPF implements [ports.Patient].
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

// FindByID implements [ports.Patient].
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
