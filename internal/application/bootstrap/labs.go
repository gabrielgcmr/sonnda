// internal/app/bootstrap/labs.go
// File: internal/app/bootstrap/labs.go
package bootstrap

import (
	labshandler "github.com/gabrielgcmr/sonnda/internal/api/handlers"
	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type LabsModule struct {
	Handler *labshandler.LabsHandler
}

func NewLabsModule(
	dbClient *postgress.Client,
	docExtractor ports.DocumentExtractorService,
	storage ports.FileStorageService,
) *LabsModule {
	patientRepo := repo.NewPatientRepository(dbClient)
	labsRepo := repo.NewLabsRepository(dbClient)

	svc := labsvc.New(patientRepo, labsRepo)
	createUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, docExtractor)
	return &LabsModule{
		Handler: labshandler.NewLabsHandler(svc, createUC, storage),
	}
}
