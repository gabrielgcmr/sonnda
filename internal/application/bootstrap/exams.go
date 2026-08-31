package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/api/handlers"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	domainai "github.com/gabrielgcmr/sonnda/internal/domain/ai"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	documenttextinfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/documenttext"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type ExamsModule struct {
	Handler *handlers.ExamsHandler
}

func NewExamsModule(
	dbClient *postgress.Client,
	docExtractor domainai.DocumentExtractorService,
	storage domainstorage.FileStorageService,
) *ExamsModule {
	patientRepo := repo.NewPatientRepository(dbClient)
	accessRepo := repo.NewPatientAccessRepository(dbClient)
	profRepo := repo.NewProfessionalRepository(dbClient)
	examsRepo := repo.NewExamsRepository(dbClient)
	labsRepo := repo.NewLabsRepository(dbClient)

	svc := examsvc.New(patientRepo, examsRepo)
	createLabUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, docExtractor)
	textExtractor := documenttextinfra.NewCommandExtractor()
	authz := authorization.New(patientRepo, accessRepo, profRepo)

	return &ExamsModule{
		Handler: handlers.NewExams(svc, createLabUC, storage, textExtractor, authz),
	}
}
