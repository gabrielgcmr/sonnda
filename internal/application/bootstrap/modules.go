// internal/application/bootstrap/modules.go
// internal/application/bootstrap/modules.go
// File: internal/app/bootstrap/modules.go
package bootstrap

import (
	domainai "github.com/gabrielgcmr/sonnda/internal/domain/ai"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
)

type Modules struct {
	User    *UserModule
	Patient *PatientModule
	Labs    *LabsModule
}

func NewModules(
	dbClient *postgress.Client,
	docExtractor domainai.DocumentExtractorService,
	storage domainstorage.FileStorageService,
) *Modules {
	return &Modules{
		User:    NewUserModule(dbClient),
		Patient: NewPatientModule(dbClient),
		Labs:    NewLabsModule(dbClient, docExtractor, storage),
	}
}
