// internal/adapters/outbound/persistence/postgres/repository/professional.go
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sonnda-api/internal/adapters/outbound/persistence/postgres/repository/db"

	professionalsqlc "sonnda-api/internal/adapters/outbound/persistence/postgres/sqlc/generated/professional"
	"sonnda-api/internal/domain/model/professional"
	"sonnda-api/internal/domain/ports/repository"
)

var ()

type Professional struct {
	client  *db.Client
	queries *professionalsqlc.Queries
}

var _ repository.Professional = (*Professional)(nil)

func NewProfessionalRepository(client *db.Client) repository.Professional {
	return &Professional{
		client:  client,
		queries: professionalsqlc.New(client.Pool()),
	}
}

// Create implements [repository.Professional].
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

// Update implements [repository.Professional].
func (p *Professional) Update(ctx context.Context, profile *professional.Professional) error {
	panic("unimplemented")
}

// Delete implements [repository.Professional].
func (p *Professional) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// SoftDelete implements [repository.Professional].
func (p *Professional) SoftDelete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FindByID implements [repository.Professional].
func (p *Professional) FindByID(ctx context.Context, id uuid.UUID) (*professional.Professional, error) {
	panic("unimplemented")
}

// FindByUserID implements [repository.Professional].
func (p *Professional) FindByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error) {
	panic("unimplemented")
}

// FindByRegistration implements [repository.Professional].
func (p *Professional) FindByRegistration(ctx context.Context, registrationNumber string, registrationIssuer string) (*professional.Professional, error) {
	panic("unimplemented")
}

// FindByName implements [repository.Professional].
func (p *Professional) FindByName(ctx context.Context, name string, limit int, offset int) ([]*professional.Professional, error) {
	panic("unimplemented")
}
