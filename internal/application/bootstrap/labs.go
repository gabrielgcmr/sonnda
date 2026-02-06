// internal/application/bootstrap/labs.go
package bootstrap

import (
	handlers "github.com/gabrielgcmr/sonnda/internal/api/handlers"
	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	domainai "github.com/gabrielgcmr/sonnda/internal/domain/ai"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type LabsModule struct {
	Handler *handlers.LabsHandler
}

func NewLabsModule(
	dbClient *postgress.Client,
	docExtractor domainai.DocumentExtractorService,
	storage domainstorage.FileStorageService,
) *LabsModule {
	patientRepo := repo.NewPatientRepository(dbClient)
	labsRepo := repo.NewLabsRepository(dbClient)

	svc := labsvc.New(patientRepo, labsRepo)
	createUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, docExtractor)
	return &LabsModule{
		Handler: handlers.NewLabs(svc, createUC, storage),
	}
}
