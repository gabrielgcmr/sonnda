// internal/app/bootstrap/modules.go
// File: internal/app/bootstrap/modules.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres/repository/db"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
)

type Modules struct {
	User    *UserModule
	Patient *PatientModule
	Labs    *LabsModule
}

func NewModules(
	dbClient *db.Client,
	identityService ports.IdentityService,
	docExtractor ports.DocumentExtractorService,
	storage ports.FileStorageService,
) *Modules {
	return &Modules{
		User:    NewUserModule(dbClient, identityService),
		Patient: NewPatientModule(dbClient),
		Labs:    NewLabsModule(dbClient, docExtractor, storage),
	}
}
