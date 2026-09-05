// internal/application/bootstrap/labs.go
package bootstrap

import (
	handlers "github.com/gabrielgcmr/sonnda/internal/api/handlers"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type LabsModule struct {
	Handler *handlers.LabsHandler
}

func NewLabsModule(
	dbClient *postgress.Client,
	labExtractor labextraction.LabReportExtractor,
	storage domainstorage.FileStorageService,
) *LabsModule {
	patientRepo := repo.NewPatientRepository(dbClient)
	accessRepo := repo.NewPatientAccessRepository(dbClient)
	profRepo := repo.NewProfessionalRepository(dbClient)
	labsRepo := repo.NewLabsRepository(dbClient)

	svc := labsvc.New(patientRepo, labsRepo)
	createUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, labExtractor)
	authz := authorization.New(patientRepo, accessRepo, profRepo)
	return &LabsModule{
		Handler: handlers.NewLabs(svc, createUC, storage, authz),
	}
}
