package supabase

import (
	"context"
	"errors"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"

	"github.com/jackc/pgx/v5"
)

type PatientRepository struct {
	client *Client
}

var _ repositories.PatientRepository = (*PatientRepository)(nil)

func NewPatientRepository(client *Client) repositories.PatientRepository {
	return &PatientRepository{client: client}
}

/* ==========================
   SQL CONSTANTS
   ========================== */

const (
	patientSelectBase = `SELECT id, cpf, cns, full_name, birth_date, gender, race, avatar_url, phone, created_at, updated_at
		FROM patients
	`
	patientByIDQuery = patientSelectBase + ` WHERE id=$1 LIMIT 1;`

	patientByCPFQuery = patientSelectBase + ` WHERE cpf=$1 LIMIT 1;`

	patientByUserIDQuery = patientSelectBase + ` WHERE app_user_id=$1 LIMIT 1;`

	patientListQuery = patientSelectBase + ` ORDER BY full_name ASC LIMIT $1 OFFSET $2;`

	patientInsertQuery = `
	INSERT INTO patients (
		app_user_id, cpf, cns, full_name, birth_date, gender, race, avatar_url, phone
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	RETURNING id, created_at, updated_at;
	`
	patientUpdateQuery = `
		UPDATE patients
		set cpf = $2,
		    cns = $3,
		    full_name = $4,
		    birth_date = $5,
		    gender = $6,
		    race = $7,
		    avatar_url = $8,
		    phone = $9,
		    updated_at = now()
		WHERE id = $1
		RETURNING updated_at
	`
)

/* ==========================
   HELPERS
   ========================== */

func scanPatient(row pgx.Row) (*domain.Patient, error) {
	var p domain.Patient
	if err := row.Scan(
		&p.ID,
		&p.CPF,
		&p.CNS,
		&p.FullName,
		&p.BirthDate,
		&p.Gender,
		&p.Race,
		&p.AvatarURL,
		&p.Phone,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &p, nil
}

/*
==========================

	PUBLIC METHODS
	==========================
*/
func (r *PatientRepository) Create(ctx context.Context, p *domain.Patient) error {

	err := r.client.Pool().QueryRow(ctx, patientInsertQuery,
		p.AppUserID,
		p.CPF,
		p.CNS,
		p.FullName,
		p.BirthDate,
		p.Gender,
		p.Race,
		p.AvatarURL,
		p.Phone,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	return err
}

// Update atualiza os dados do paciente.
func (r *PatientRepository) Update(ctx context.Context, p *domain.Patient) error {
	if p == nil {
		return errors.New("patient is nil")
	}

	row := r.client.Pool().QueryRow(ctx, patientUpdateQuery,
		p.ID,
		p.CPF,
		p.CNS,
		p.FullName,
		p.BirthDate,
		p.Gender,
		p.Race,
		p.AvatarURL,
		p.Phone,
	)
	if err := row.Scan(&p.UpdatedAt); err != nil {
		return err
	}
	return nil
}

// Delete remove um paciente.
func (r *PatientRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Pool().Exec(ctx, "DELETE FROM patients WHERE id=$1", id)
	return err
}

// Busca um paciente pelo user ID.
func (r *PatientRepository) FindByUserID(ctx context.Context, userID string) (*domain.Patient, error) {
	row := r.client.Pool().QueryRow(ctx, patientByUserIDQuery, userID)
	patient, err := scanPatient(row)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return patient, nil
}

// FindByCPF busca um paciente pelo CPF.
func (r *PatientRepository) FindByCPF(ctx context.Context, cpf string) (*domain.Patient, error) {
	row := r.client.Pool().QueryRow(ctx, patientByCPFQuery, cpf)
	patient, err := scanPatient(row)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return patient, nil
}

func (r *PatientRepository) FindByID(ctx context.Context, id string) (*domain.Patient, error) {
	row := r.client.Pool().QueryRow(ctx, patientByIDQuery, id)
	patient, err := scanPatient(row)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return patient, nil
}

// List retorna pacientes paginados.
func (r *PatientRepository) List(ctx context.Context, limit, offset int) ([]domain.Patient, error) {
	rows, err := r.client.Pool().Query(ctx, patientListQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Patient
	for rows.Next() {
		p, err := scanPatient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}

	return out, nil
}
