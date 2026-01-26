// internal/adapters/outbound/data/postgres/repository/professional.go
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/gabrielgcmr/sonnda/internal/adapters/outbound/data/postgres/repository/db"

	professionalsqlc "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/data/postgres/sqlc/generated/professional"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/professional"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
)

type Professional struct {
	client  *db.Client
	queries *professionalsqlc.Queries
}

var _ ports.ProfessionalRepo = (*Professional)(nil)

func NewProfessionalRepository(client *db.Client) ports.ProfessionalRepo {
	return &Professional{
		client:  client,
		queries: professionalsqlc.New(client.Pool()),
	}
}

// Create implements [ports.ProfessionalRepo].
func (p *Professional) Create(ctx context.Context, prof *professional.Professional) error {
	if prof == nil {
		return errors.New("profile is nil")
	}

	row, err := p.queries.CreateProfessional(ctx, professionalsqlc.CreateProfessionalParams{
		UserID:             prof.UserID,
		Kind:               string(prof.Kind),
		RegistrationNumber: prof.RegistrationNumber,
		RegistrationIssuer: prof.RegistrationIssuer,
		RegistrationState:  FromNullableStringToPgText(prof.RegistrationState),
		Status:             string(prof.Status),
	})
	if err != nil {
		if IsUniqueViolationError(err) {
			return ErrProfessionalAlreadyExists
		}
		return errors.Join(ErrRepositoryFailure, err)
	}

	prof.UserID = row.UserID
	prof.Kind = professional.Kind(row.Kind)
	prof.RegistrationNumber = row.RegistrationNumber
	prof.RegistrationIssuer = row.RegistrationIssuer
	prof.RegistrationState = FromPgTextToNullableString(row.RegistrationState)
	prof.Status = professional.VerificationStatus(row.Status)
	prof.VerifiedAt = FromPgTimestamptzToNullableTimestamptz(row.VerifiedAt)
	prof.CreatedAt = row.CreatedAt.Time
	prof.UpdatedAt = row.UpdatedAt.Time

	return nil
}

// Update implements [ports.ProfessionalRepo].
func (p *Professional) Update(ctx context.Context, profile *professional.Professional) error {
	panic("unimplemented")
}

// Delete implements [ports.ProfessionalRepo].
func (p *Professional) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// SoftDelete implements [ports.ProfessionalRepo].
func (p *Professional) SoftDelete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FindByID implements [ports.ProfessionalRepo].
func (p *Professional) FindByID(ctx context.Context, id uuid.UUID) (*professional.Professional, error) {
	panic("unimplemented")
}

// FindByUserID implements [ports.ProfessionalRepo].
func (p *Professional) FindByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error) {
	panic("unimplemented")
}

// FindByRegistration implements [ports.ProfessionalRepo].
func (p *Professional) FindByRegistration(ctx context.Context, registrationNumber string, registrationIssuer string) (*professional.Professional, error) {
	panic("unimplemented")
}

// FindByName implements [ports.ProfessionalRepo].
func (p *Professional) FindByName(ctx context.Context, name string, limit int, offset int) ([]*professional.Professional, error) {
	panic("unimplemented")
}
