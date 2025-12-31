package patient

import (
	"context"
	"time"

	"sonnda-api/internal/domain/entities/patient"
	"sonnda-api/internal/domain/entities/shared"
	"sonnda-api/internal/domain/ports/repositories"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
	"sonnda-api/internal/infrastructure/persistence/repository/helpers"
	patientsqlc "sonnda-api/internal/infrastructure/persistence/sqlc/generated/patient"

	"github.com/google/uuid"
)

type PatientRepository struct {
	client  *db.Client
	queries *patientsqlc.Queries
}

var _ repositories.PatientRepository = (*PatientRepository)(nil)

func NewPatientRepository(client *db.Client) repositories.PatientRepository {
	return &PatientRepository{
		client:  client,
		queries: patientsqlc.New(client.Pool()),
	}
}

// Create implements [repositories.PatientRepository].
func (r *PatientRepository) Create(ctx context.Context, p *patient.Patient) error {
	params := patientsqlc.CreatePatientParams{
		ID:        p.ID,
		UserID:    helpers.FromNullableUUIDToPgUUID(p.UserID),
		Cpf:       p.CPF,
		Cns:       helpers.FromNullableStringToPgText(p.CNS),
		FullName:  p.FullName,
		BirthDate: helpers.FromRequiredDateToPgDate(p.BirthDate),
		Gender:    string(p.Gender),
		Race:      string(p.Race),
		Phone:     helpers.FromNullableStringToPgText(p.Phone),
		AvatarUrl: helpers.FromNullableStringToPgText(&p.AvatarURL),
	}

	_, err := r.queries.CreatePatient(ctx, params)
	return err
}

// SoftDelete implements [repositories.PatientRepository].
func (p *PatientRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FindByCPF implements [repositories.PatientRepository].
func (p *PatientRepository) FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	row, err := p.queries.GetPatientByCPF(ctx, cpf)
	if err != nil {
		if helpers.IsPgNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	userID, err := helpers.FromPgUUIDToNullableUUID(row.UserID)
	if err != nil {
		return nil, err
	}

	return &patient.Patient{
		ID:        row.ID,
		UserID:    userID,
		CPF:       row.Cpf,
		CNS:       helpers.FromPgTextToNullableString(row.Cns),
		FullName:  row.FullName,
		BirthDate: row.BirthDate.Time,
		Gender:    shared.Gender(row.Gender),
		Race:      shared.Race(row.Race),
		AvatarURL: row.AvatarUrl.String,
		Phone:     helpers.FromPgTextToNullableString(row.Phone),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

// FindByID implements [repositories.PatientRepository].
func (p *PatientRepository) FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error) {
	row, err := p.queries.GetPatientByID(ctx, id)
	if err != nil {
		if helpers.IsPgNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	userID, err := helpers.FromPgUUIDToNullableUUID(row.UserID)
	if err != nil {
		return nil, err
	}

	return &patient.Patient{
		ID:        row.ID,
		UserID:    userID,
		CPF:       row.Cpf,
		CNS:       helpers.FromPgTextToNullableString(row.Cns),
		FullName:  row.FullName,
		BirthDate: row.BirthDate.Time,
		Gender:    shared.Gender(row.Gender),
		Race:      shared.Race(row.Race),
		AvatarURL: row.AvatarUrl.String,
		Phone:     helpers.FromPgTextToNullableString(row.Phone),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

// List implements [repositories.PatientRepository].
func (p *PatientRepository) List(ctx context.Context, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// ListByBirthDate implements [repositories.PatientRepository].
func (p *PatientRepository) ListByBirthDate(ctx context.Context, birthDate time.Time, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// ListByIDs implements [repositories.PatientRepository].
func (p *PatientRepository) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]patient.Patient, error) {
	panic("unimplemented")
}

// ListByName implements [repositories.PatientRepository].
func (p *PatientRepository) ListByName(ctx context.Context, name string, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// Update implements [repositories.PatientRepository].
func (p *PatientRepository) Update(ctx context.Context, patient *patient.Patient) error {
	panic("unimplemented")
}
