// internal/app/bootstrap/modules.go
// File: internal/app/bootstrap/modules.go
package bootstrap

import (
	"sonnda-api/internal/adapters/outbound/persistence/postgres/repository/db"
	"sonnda-api/internal/domain/ports/integration"
	"sonnda-api/internal/domain/ports/integration/documentai"
)

type Modules struct {
	User    *UserModule
	Patient *PatientModule
	Labs    *LabsModule
}

func NewModules(
	dbClient *db.Client,
	identityService integration.IdentityService,
	docExtractor documentai.DocumentExtractor,
	storage integration.StorageService,
) *Modules {
	return &Modules{
		User:    NewUserModule(dbClient, identityService),
		Patient: NewPatientModule(dbClient),
		Labs:    NewLabsModule(dbClient, docExtractor, storage),
	}
}
