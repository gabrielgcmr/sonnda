// internal/app/bootstrap/labs.go
// File: internal/app/bootstrap/labs.go
package bootstrap

import (
	labshandler "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/handlers"
	postgress "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres"
	repo "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres/repository"
	labsvc "github.com/gabrielgcmr/sonnda/internal/app/services/labs"
	labsuc "github.com/gabrielgcmr/sonnda/internal/app/usecase/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
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
