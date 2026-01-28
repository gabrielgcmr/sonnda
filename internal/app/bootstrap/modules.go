// internal/app/bootstrap/modules.go
// File: internal/app/bootstrap/modules.go
package bootstrap

import (
	postgress "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

type Modules struct {
	User    *UserModule
	Patient *PatientModule
	Labs    *LabsModule
}

func NewModules(
	dbClient *postgress.Client,
	identityProvider security.IdentityProvider,
	docExtractor ports.DocumentExtractorService,
	storage ports.FileStorageService,
) *Modules {
	return &Modules{
		User:    NewUserModule(dbClient, identityProvider),
		Patient: NewPatientModule(dbClient),
		Labs:    NewLabsModule(dbClient, docExtractor, storage),
	}
}
