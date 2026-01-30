// internal/app/bootstrap/modules.go
// File: internal/app/bootstrap/modules.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
)

type Modules struct {
	User    *UserModule
	Patient *PatientModule
	Labs    *LabsModule
}

func NewModules(
	dbClient *postgress.Client,
	docExtractor ports.DocumentExtractorService,
	storage ports.FileStorageService,
) *Modules {
	return &Modules{
		User:    NewUserModule(dbClient),
		Patient: NewPatientModule(dbClient),
		Labs:    NewLabsModule(dbClient, docExtractor, storage),
	}
}
