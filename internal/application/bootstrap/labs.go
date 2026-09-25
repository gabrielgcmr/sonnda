// internal/application/bootstrap/labs.go
package bootstrap

import (
	handlers "github.com/gabrielgcmr/sonnda/internal/api/handlers"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	processingpostgres "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/postgres"
	labsuc "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/processing"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	accesspostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/access/postgres"
	labsvc "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	laboratoryhttp "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/http"
	labpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/postgres"
	patientpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/postgres"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
)

type LabsModule struct {
	Handler           *handlers.LabsHandler
	LaboratoryHandler *laboratoryhttp.Handler
}

func NewLabsModule(
	dbClient *postgress.Client,
	labExtractor labextraction.LabReportExtractor,
	storage domainstorage.FileStorageService,
) *LabsModule {
	patientRepo := patientpostgres.NewRepository(dbClient)
	accessRepo := accesspostgres.NewRepository(dbClient)
	labsRepo := processingpostgres.NewRepository(dbClient, labpostgres.NewLabsRepository(dbClient))

	svc := labsvc.New(patientRepo, labsRepo)
	createUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, labExtractor)
	accessChecker := patientaccess.NewChecker(patientRepo, accessRepo)
	return &LabsModule{
		Handler:           handlers.NewLabs(createUC, storage, accessChecker),
		LaboratoryHandler: laboratoryhttp.NewHandler(svc, accessChecker),
	}
}
