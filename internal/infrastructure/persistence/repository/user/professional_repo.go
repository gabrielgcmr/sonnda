package user

import (
	"context"

	"sonnda-api/internal/domain/model/user/professional"
	"sonnda-api/internal/domain/ports/repositories"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
	professionalsqlc "sonnda-api/internal/infrastructure/persistence/sqlc/generated/professional"

	"github.com/google/uuid"
)

type Professional struct {
	client  *db.Client
	queries *professionalsqlc.Queries
}

var _ repositories.ProfessionalRepository = (*Professional)(nil)

func NewProfessionalRepository(client *db.Client) repositories.ProfessionalRepository {
	return &Professional{
		client:  client,
		queries: professionalsqlc.New(client.Pool()),
	}
}

// Create implements [repositories.ProfessionalRepository].
func (p *Professional) Create(ctx context.Context, profile *professional.Professional) error {
	panic("unimplemented")
}

// Delete implements [repositories.ProfessionalRepository].
func (p *Professional) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FindByID implements [repositories.ProfessionalRepository].
func (p *Professional) FindByID(ctx context.Context, id uuid.UUID) (*professional.Professional, error) {
	panic("unimplemented")
}

// FindByName implements [repositories.ProfessionalRepository].
func (p *Professional) FindByName(ctx context.Context, name string, limit int, offset int) ([]*professional.Professional, error) {
	panic("unimplemented")
}

// FindByRegistration implements [repositories.ProfessionalRepository].
func (p *Professional) FindByRegistration(ctx context.Context, registrationNumber string, registrationIssuer string) (*professional.Professional, error) {
	panic("unimplemented")
}

// FindByUserID implements [repositories.ProfessionalRepository].
func (p *Professional) FindByUserID(ctx context.Context, userID uuid.UUID) (*professional.Professional, error) {
	panic("unimplemented")
}

// Update implements [repositories.ProfessionalRepository].
func (p *Professional) Update(ctx context.Context, profile *professional.Professional) error {
	panic("unimplemented")
}
