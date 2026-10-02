// internal/features/patient/exam/laboratory/postgres/repository.go
package postgres

import (
	labrepository "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres"
	labsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres/sqlc/generated/lab"
)

// Repository is the PostgreSQL adapter for laboratory reports.
type Repository struct {
	client  *postgress.Client
	queries *labsqlc.Queries
}

var _ labrepository.Repository = (*Repository)(nil)

// NewRepository creates a laboratory repository backed by the shared PostgreSQL client.
func NewRepository(client *postgress.Client) *Repository {
	return &Repository{
		client:  client,
		queries: labsqlc.New(client.Pool()),
	}
}
