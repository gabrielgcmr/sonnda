package bootstrap

import (
	"sonnda-api/internal/app/interfaces/external"
	labsvc "sonnda-api/internal/app/services/labs"
	labshandler "sonnda-api/internal/http/api/handlers/labs"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
	labsrepo "sonnda-api/internal/infrastructure/persistence/repository/labs"
	patientrepo "sonnda-api/internal/infrastructure/persistence/repository/patient"
)

func NewLabsModule(
	dbClient *db.Client,
	docExtractor external.DocumentExtractor,
	storage external.StorageService,
) *labshandler.LabsHandler {
	patientRepo := patientrepo.NewPatientRepository(dbClient)
	labsRepo := labsrepo.NewLabsRepository(dbClient)

	svc := labsvc.New(patientRepo, labsRepo, docExtractor)
	return labshandler.NewLabsHandler(svc, storage)
}
