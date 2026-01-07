package bootstrap

import (
	labsvc "sonnda-api/internal/app/services/labs"
	labshandler "sonnda-api/internal/adapters/inbound/http/api/handlers/labs"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	labsrepo "sonnda-api/internal/adapters/outbound/persistence/repository/labs"
	patientrepo "sonnda-api/internal/adapters/outbound/persistence/repository/patient"
	"sonnda-api/internal/domain/ports/integrations"
)

func NewLabsModule(
	dbClient *db.Client,
	docExtractor integrations.DocumentExtractor,
	storage integrations.StorageService,
) *labshandler.LabsHandler {
	patientRepo := patientrepo.NewPatientRepository(dbClient)
	labsRepo := labsrepo.NewLabsRepository(dbClient)

	svc := labsvc.New(patientRepo, labsRepo, docExtractor)
	return labshandler.NewLabsHandler(svc, storage)
}
