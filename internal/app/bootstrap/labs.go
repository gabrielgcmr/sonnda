package bootstrap

import (
	labshandler "sonnda-api/internal/adapters/inbound/http/api/handlers/labs"
	labsrepo "sonnda-api/internal/adapters/outbound/persistence/repository"
	patientrepo "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	labsvc "sonnda-api/internal/app/services/labs"
	"sonnda-api/internal/domain/ports/integration"
)

func NewLabsModule(
	dbClient *db.Client,
	docExtractor integration.DocumentExtractor,
	storage integration.StorageService,
) *labshandler.LabsHandler {
	patientRepo := patientrepo.NewPatientRepository(dbClient)
	labsRepo := labsrepo.NewLabsRepository(dbClient)

	svc := labsvc.New(patientRepo, labsRepo, docExtractor)
	return labshandler.NewLabsHandler(svc, storage)
}
