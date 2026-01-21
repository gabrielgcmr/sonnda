// File: internal/app/bootstrap/labs.go
package bootstrap

import (
	labshandler "sonnda-api/internal/adapters/inbound/http/api/handlers"
	repo "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	labsvc "sonnda-api/internal/app/services/labs"
	labsuc "sonnda-api/internal/app/usecase/labs"
	"sonnda-api/internal/domain/ports/integration"
	"sonnda-api/internal/domain/ports/integration/documentai"
)

type LabsModule struct {
	Handler *labshandler.LabsHandler
}

func NewLabsModule(
	dbClient *db.Client,
	docExtractor documentai.DocumentExtractor,
	storage integration.StorageService,
) *LabsModule {
	patientRepo := repo.NewPatientRepository(dbClient)
	labsRepo := repo.NewLabsRepository(dbClient)

	svc := labsvc.New(patientRepo, labsRepo)
	createUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, docExtractor)
	return &LabsModule{
		Handler: labshandler.NewLabsHandler(svc, createUC, storage),
	}
}
