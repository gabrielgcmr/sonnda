// internal/features/patient/profile/postgres/repository.go
package patientpostgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/demographics"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patientaccess"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	patientsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/patient"
	patientaccesssqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/patientaccess"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository struct {
	client  *postgress.Client
	queries *patientsqlc.Queries
}

// FindByName implements [patientprofile.Repository].
func (r *Repository) FindByName(ctx context.Context, name string) ([]patient.Patient, error) {
	panic("unimplemented")
}

// SearchByName implements [patientprofile.Repository].
func (r *Repository) SearchByName(ctx context.Context, name string, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// HardDelete implements [patientprofile.Repository].
func (r *Repository) HardDelete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// List implements [patientprofile.Repository].
func (r *Repository) List(ctx context.Context, limit int, offset int) ([]patient.Patient, error) {
	panic("unimplemented")
}

// Update implements [patientprofile.Repository].
func (r *Repository) Update(ctx context.Context, patient *patient.Patient) error {
	panic("unimplemented")
}

var _ patientprofile.Repository = (*Repository)(nil)

func NewRepository(client *postgress.Client) patientprofile.Repository {
	return &Repository{
		client:  client,
		queries: patientsqlc.New(client.Pool()),
	}
}

// Create implements [patientprofile.Repository].
func (r *Repository) Create(ctx context.Context, p *patient.Patient) error {
	return r.createWithQueries(ctx, r.queries, p)
}

// CreateWithAccess creates a patient and its initial access grant atomically.
func (r *Repository) CreateWithAccess(
	ctx context.Context,
	p *patient.Patient,
	access *patientaccess.PatientAccess,
) error {
	if err := access.Validate(); err != nil {
		return fmt.Errorf("invalid patient access: %w", err)
	}

	tx, err := r.client.BeginTx(ctx)
	if err != nil {
		return errors.Join(repository.ErrRepositoryFailure, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	patientQueries := r.queries.WithTx(tx)
	accessQueries := patientaccesssqlc.New(r.client.Pool()).WithTx(tx)

	if err := r.createWithQueries(ctx, patientQueries, p); err != nil {
		return err
	}

	var grantedBy pgtype.UUID
	if access.GrantedBy != nil {
		grantedBy = pgtype.UUID{Bytes: *access.GrantedBy, Valid: true}
	}

	err = accessQueries.UpsertPatientAccess(ctx, patientaccesssqlc.UpsertPatientAccessParams{
		PatientID:    pgtype.UUID{Bytes: access.PatientID, Valid: true},
		GranteeID:    pgtype.UUID{Bytes: access.GranteeID, Valid: true},
		RelationType: string(access.RelationType),
		GrantedBy:    grantedBy,
	})
	if err != nil {
		return errors.Join(repository.ErrRepositoryFailure, fmt.Errorf("failed to upsert patient access: %w", err))
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Join(repository.ErrRepositoryFailure, err)
	}

	return nil
}

func (r *Repository) createWithQueries(
	ctx context.Context,
	queries *patientsqlc.Queries,
	p *patient.Patient,
) error {
	params := patientsqlc.CreatePatientParams{
		ID:          p.ID,
		OwnerUserID: fromNullableUUID(p.OwnerUserID),
		Cpf:         p.CPF,
		Cns:         fromNullableString(p.CNS),
		FullName:    p.FullName,
		BirthDate:   pgtype.Date{Time: p.BirthDate, Valid: true},
		Gender:      string(p.Gender),
		Race:        string(p.Race),
		Phone:       fromNullableString(p.Phone),
		AvatarUrl:   fromNullableString(&p.AvatarURL),
	}

	row, err := queries.CreatePatient(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			return patientprofile.ErrPatientAlreadyExists
		}
		return errors.Join(repository.ErrRepositoryFailure, err)
	}

	p.ID = row.ID
	p.OwnerUserID = fromPgUUID(row.OwnerUserID)
	p.CPF = row.Cpf
	p.CNS = fromPgText(row.Cns)
	p.FullName = row.FullName
	p.BirthDate = row.BirthDate.Time
	p.Gender = demographics.Gender(row.Gender)
	p.Race = demographics.Race(row.Race)
	p.AvatarURL = row.AvatarUrl.String
	p.Phone = fromPgText(row.Phone)
	p.CreatedAt = row.CreatedAt.Time
	p.UpdatedAt = row.UpdatedAt.Time

	return nil
}

// SoftDelete implements [patientprofile.Repository].
func (p *Repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FindByCPF implements [patientprofile.Repository].
func (p *Repository) FindByCPF(ctx context.Context, cpf string) (*patient.Patient, error) {
	row, err := p.queries.GetPatientByCPF(ctx, cpf)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Join(repository.ErrRepositoryFailure, err)
	}

	return &patient.Patient{
		ID:          row.ID,
		OwnerUserID: fromPgUUID(row.OwnerUserID),
		CPF:         row.Cpf,
		CNS:         fromPgText(row.Cns),
		FullName:    row.FullName,
		BirthDate:   row.BirthDate.Time,
		Gender:      demographics.Gender(row.Gender),
		Race:        demographics.Race(row.Race),
		AvatarURL:   row.AvatarUrl.String,
		Phone:       fromPgText(row.Phone),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

// FindByID implements [patientprofile.Repository].
func (p *Repository) FindByID(ctx context.Context, id uuid.UUID) (*patient.Patient, error) {
	row, err := p.queries.GetPatientByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Join(repository.ErrRepositoryFailure, err)
	}

	return &patient.Patient{
		ID:          row.ID,
		OwnerUserID: fromPgUUID(row.OwnerUserID),
		CPF:         row.Cpf,
		CNS:         fromPgText(row.Cns),
		FullName:    row.FullName,
		BirthDate:   row.BirthDate.Time,
		Gender:      demographics.Gender(row.Gender),
		Race:        demographics.Race(row.Race),
		AvatarURL:   row.AvatarUrl.String,
		Phone:       fromPgText(row.Phone),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func fromNullableString(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func fromPgText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func fromNullableUUID(value *uuid.UUID) pgtype.UUID {
	if value == nil || *value == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *value, Valid: true}
}

func fromPgUUID(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	result := uuid.UUID(value.Bytes)
	return &result
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
