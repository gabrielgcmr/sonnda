// internal/adapters/outbound/storage/postgres/repository/patient_access.go
package repo

import (
	"context"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/domain/model/patientaccess"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	patientaccesssqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/patientaccess"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type PatientAccessRepository struct {
	client  *postgress.Client
	queries *patientaccesssqlc.Queries
}

var _ ports.PatientAccessRepo = (*PatientAccessRepository)(nil)

func NewPatientAccessRepository(client *postgress.Client) ports.PatientAccessRepo {
	return &PatientAccessRepository{
		client:  client,
		queries: patientaccesssqlc.New(client.Pool()),
	}
}

// ListAccessiblePatientsByUser implements [ports.PatientAccessRepo].
func (p *PatientAccessRepository) ListAccessiblePatientsByUser(ctx context.Context, granteeID uuid.UUID, limit, offset int) ([]ports.AccessiblePatient, int64, error) {
	// Buscar lista paginada
	rows, err := p.queries.ListAccessiblePatientsByUser(ctx, patientaccesssqlc.ListAccessiblePatientsByUserParams{
		GranteeID: pgtype.UUID{Bytes: granteeID, Valid: true},
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list accessible patients: %w", err)
	}

	// Buscar count total
	total, err := p.queries.CountAccessiblePatientsByUser(ctx, pgtype.UUID{Bytes: granteeID, Valid: true})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count accessible patients: %w", err)
	}

	// Mapear para DTO
	result := make([]ports.AccessiblePatient, len(rows))
	for i, row := range rows {
		var avatarURL *string
		if row.AvatarUrl.Valid {
			avatarURL = &row.AvatarUrl.String
		}

		result[i] = ports.AccessiblePatient{
			PatientID:    row.PatientID.Bytes,
			FullName:     row.FullName,
			AvatarURL:    avatarURL,
			RelationType: row.RelationType,
		}
	}

	return result, total, nil
}

// HasActiveAccess implements [ports.PatientAccessRepo].
func (p *PatientAccessRepository) HasActiveAccess(ctx context.Context, patientID uuid.UUID, granteeID uuid.UUID) (bool, error) {
	access, err := p.queries.FindPatientAccess(ctx, patientaccesssqlc.FindPatientAccessParams{
		PatientID: pgtype.UUID{Bytes: patientID, Valid: true},
		GranteeID: pgtype.UUID{Bytes: granteeID, Valid: true},
	})
	if err != nil {
		// Se não encontrou, retorna false sem erro
		return false, nil
	}

	// Verifica se não está revogado
	return !access.RevokedAt.Valid, nil
}

// Upsert implements [ports.PatientAccessRepo].
func (p *PatientAccessRepository) Upsert(ctx context.Context, access *patientaccess.PatientAccess) error {
	if err := access.Validate(); err != nil {
		return fmt.Errorf("invalid patient access: %w", err)
	}

	var grantedBy pgtype.UUID
	if access.GrantedBy != nil {
		grantedBy = pgtype.UUID{Bytes: *access.GrantedBy, Valid: true}
	}

	err := p.queries.UpsertPatientAccess(ctx, patientaccesssqlc.UpsertPatientAccessParams{
		PatientID:    pgtype.UUID{Bytes: access.PatientID, Valid: true},
		GranteeID:    pgtype.UUID{Bytes: access.GranteeID, Valid: true},
		RelationType: string(access.RelationType),
		GrantedBy:    grantedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert patient access: %w", err)
	}

	return nil
}
