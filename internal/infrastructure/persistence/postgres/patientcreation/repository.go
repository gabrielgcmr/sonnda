// internal/infrastructure/persistence/postgres/patientcreation/repository.go
package patientcreationpostgres

import (
	"context"
	"errors"
	"fmt"

	patientcreation "github.com/gabrielgcmr/sonnda/internal/application/usecase/patientcreation"
	domainrepository "github.com/gabrielgcmr/sonnda/internal/domain/repository"
	accessdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/access/domain"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"
	postgres "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	patientsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/patient"
	patientaccesssqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/patientaccess"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository struct {
	client *postgres.Client
}

func NewRepository(client *postgres.Client) patientcreation.Repository {
	return &Repository{client: client}
}

func (r *Repository) CreateWithInitialAccess(
	ctx context.Context,
	patient *profiledomain.Patient,
	access *accessdomain.PatientAccess,
) error {
	if r == nil || r.client == nil {
		return errors.Join(domainrepository.ErrRepositoryFailure, errors.New("postgres client not configured"))
	}
	if err := access.Validate(); err != nil {
		return fmt.Errorf("invalid patient access: %w", err)
	}

	tx, err := r.client.BeginTx(ctx)
	if err != nil {
		return errors.Join(domainrepository.ErrRepositoryFailure, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	patientQueries := patientsqlc.New(tx)
	accessQueries := patientaccesssqlc.New(tx)

	row, err := patientQueries.CreatePatient(ctx, patientsqlc.CreatePatientParams{
		ID:          patient.ID,
		OwnerUserID: nullableUUID(patient.OwnerUserID),
		Cpf:         patient.CPF,
		Cns:         nullableText(patient.CNS),
		FullName:    patient.FullName,
		BirthDate:   pgtype.Date{Time: patient.BirthDate, Valid: true},
		Gender:      string(patient.Gender),
		Race:        string(patient.Race),
		Phone:       nullableText(patient.Phone),
		AvatarUrl:   optionalText(patient.AvatarURL),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return patientprofile.ErrPatientAlreadyExists
		}
		return errors.Join(domainrepository.ErrRepositoryFailure, err)
	}

	patient.ID = row.ID
	patient.CreatedAt = row.CreatedAt.Time
	patient.UpdatedAt = row.UpdatedAt.Time

	if err := accessQueries.UpsertPatientAccess(ctx, patientaccesssqlc.UpsertPatientAccessParams{
		PatientID:    pgtype.UUID{Bytes: access.PatientID, Valid: true},
		GranteeID:    pgtype.UUID{Bytes: access.GranteeID, Valid: true},
		RelationType: string(access.RelationType),
		GrantedBy:    nullableUUID(access.GrantedBy),
	}); err != nil {
		return errors.Join(domainrepository.ErrRepositoryFailure, fmt.Errorf("upsert initial patient access: %w", err))
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Join(domainrepository.ErrRepositoryFailure, err)
	}
	return nil
}

func nullableUUID(value *uuid.UUID) pgtype.UUID {
	if value == nil || *value == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *value, Valid: true}
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

var _ patientcreation.Repository = (*Repository)(nil)
