package bootstrap

import (
	labshandler "sonnda-api/internal/adapters/inbound/http/api/handlers"
	labsrepo "sonnda-api/internal/adapters/outbound/persistence/repository"
	patientrepo "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	labsvc "sonnda-api/internal/app/services/labs"
	labsuc "sonnda-api/internal/app/usecase/labs"
	"sonnda-api/internal/domain/ports/integration"
	"sonnda-api/internal/domain/ports/integration/documentai"
)

func NewLabsModule(
	dbClient *db.Client,
	docExtractor documentai.DocumentExtractor,
	storage integration.StorageService,
) *labshandler.LabsHandler {
	patientRepo := patientrepo.NewPatientRepository(dbClient)
	labsRepo := labsrepo.NewLabsRepository(dbClient)

	svc := labsvc.New(patientRepo, labsRepo)
	createUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, docExtractor)
	return labshandler.NewLabsHandler(svc, createUC, storage)
}
