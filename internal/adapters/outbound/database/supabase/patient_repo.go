package supabase

import (
	"context"
	"errors"
	"fmt"

	patientssqlc "sonnda-api/internal/adapters/outbound/database/sqlc/patients"
	"sonnda-api/internal/core/domain/demographics"
	"sonnda-api/internal/core/domain/patient"
	"sonnda-api/internal/core/ports/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PatientRepository struct {
	client  *Client
	queries *patientssqlc.Queries
}

var _ repositories.PatientRepository = (*PatientRepository)(nil)

func NewPatientRepository(client *Client) repositories.PatientRepository {
	return &PatientRepository{
		client:  client,
		queries: patientssqlc.New(client.Pool()),
	}
}

func (r *PatientRepository) Create(ctx context.Context, p *patient.Patient) error {
	if p == nil {
		return errors.New("patient is nil")
	}

	if p.ID == "" {
		p.ID = uuid.NewString()
	}

	dbPatient, err := r.queries.CreatePatient(ctx, patientssqlc.CreatePatientParams{
		ID:        p.ID,
		AppUserID: ToText(p.AppUserID),
		Cpf:       p.CPF,
		Cns:       ToText(p.CNS),
		FullName:  p.FullName,
		BirthDate: pgtype.Date{Time: p.BirthDate, Valid: true},
		Gender:    string(p.Gender),
		Race:      string(p.Race),
		Phone:     ToText(p.Phone),
		AvatarUrl: p.AvatarURL,
	})
	if err != nil {
		return err
	}

	created, err := dbPatientToDomain(dbPatient)
	if err != nil {
		return err
	}

	*p = *created

	return nil
}

// Update atualiza os dados do paciente.
func (r *PatientRepository) Update(ctx context.Context, p *patient.Patient) error {
	if p == nil {
		return errors.New("patient is nil")
	}

	dbPatient, err := r.queries.UpdatePatient(ctx, patientssqlc.UpdatePatientParams{
		ID:        p.ID,
		Cpf:       p.CPF,
		Cns:       ToText(p.CNS),
		FullName:  p.FullName,
		BirthDate: pgtype.Date{Time: p.BirthDate, Valid: true},
		Gender:    string(p.Gender),
		Race:      string(p.Race),
		Phone:     ToText(p.Phone),
		AvatarUrl: p.AvatarURL,
	})
	if err != nil {
		return err
	}

	updated, err := dbPatientToDomain(dbPatient)
	if err != nil {
		return err
	}

	*p = *updated

	return nil
}

// Delete remove um paciente.
func (r *PatientRepository) Delete(ctx context.Context, id string) error {
	_, err := r.queries.DeletePatient(ctx, id)
	return err
}

// Busca um paciente pelo user ID.
func (r *PatientRepository) FindByUserID(ctx context.Context, userID string) (*patient.Patient, error) {
	dbPatient, err := r.queries.FindPatientByUserID(ctx, ToTextValue(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return dbPatientToDomain(dbPatient)
}

// FindByCPF busca um paciente pelo CPF.
func (r *PatientRepository) FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	dbPatient, err := r.queries.FindPatientByCPF(ctx, cpf)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return dbPatientToDomain(dbPatient)
}

func (r *PatientRepository) FindByID(ctx context.Context, id string) (*patient.Patient, error) {
	dbPatient, err := r.queries.FindPatientByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return dbPatientToDomain(dbPatient)
}

// List retorna pacientes paginados.
func (r *PatientRepository) List(ctx context.Context, limit, offset int) ([]patient.Patient, error) {
	rows, err := r.queries.ListPatients(ctx, patientssqlc.ListPatientsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]patient.Patient, 0, len(rows))
	for _, row := range rows {
		patient, err := dbPatientToDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, *patient)
	}

	return out, nil
}

func dbPatientToDomain(p patientssqlc.Patient) (*patient.Patient, error) {
	if p.ID == "" {
		return nil, fmt.Errorf("patient id is empty")
	}
	if !p.BirthDate.Valid {
		return nil, fmt.Errorf("patient birth_date is null")
	}

	createdAt, err := MustTime(p.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := MustTime(p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	birthDate := p.BirthDate.Time

	return &patient.Patient{
		ID:        p.ID,
		AppUserID: FromText(p.AppUserID),
		CPF:       p.Cpf,
		CNS:       FromText(p.Cns),
		FullName:  p.FullName,
		BirthDate: birthDate,
		Gender:    demographics.Gender(p.Gender),
		Race:      demographics.Race(p.Race),
		AvatarURL: p.AvatarUrl,
		Phone:     FromText(p.Phone),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
