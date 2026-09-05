// internal/application/bootstrap/exams.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/api/handlers"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
	textextractioninfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/textextraction"
)

type ExamsModule struct {
	Handler *handlers.ExamsHandler
}

func NewExamsModule(
	dbClient *postgress.Client,
	labExtractor labextraction.LabReportExtractor,
	storage domainstorage.FileStorageService,
) *ExamsModule {
	patientRepo := repo.NewPatientRepository(dbClient)
	accessRepo := repo.NewPatientAccessRepository(dbClient)
	profRepo := repo.NewProfessionalRepository(dbClient)
	examsRepo := repo.NewExamsRepository(dbClient)
	labsRepo := repo.NewLabsRepository(dbClient)

	svc := examsvc.New(patientRepo, examsRepo)
	createLabUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, labExtractor)
	textExtractor := textextractioninfra.NewCommandExtractor()
	authz := authorization.New(patientRepo, accessRepo, profRepo)

	return &ExamsModule{
		Handler: handlers.NewExams(svc, createLabUC, storage, textExtractor, authz),
	}
}
