package bootstrap

import (
	labsvc "sonnda-api/internal/app/services/labs"
	"sonnda-api/internal/domain/ports/integrations"
	labshandler "sonnda-api/internal/http/api/handlers/labs"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
	labsrepo "sonnda-api/internal/infrastructure/persistence/repository/labs"
	patientrepo "sonnda-api/internal/infrastructure/persistence/repository/patient"
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
