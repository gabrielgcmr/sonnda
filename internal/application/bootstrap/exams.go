// internal/application/bootstrap/exams.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/api/handlers"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	textsvc "github.com/gabrielgcmr/sonnda/internal/application/services/textextraction"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	patientpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/postgres"
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
	ocrConfig config.OCRConfig,
	fallback domaintext.Extractor,
) *ExamsModule {
	patientRepo := patientpostgres.NewRepository(dbClient)
	accessRepo := repo.NewPatientAccessRepository(dbClient)
	examsRepo := repo.NewExamsRepository(dbClient)
	labsRepo := repo.NewLabsRepository(dbClient)

	svc := examsvc.New(patientRepo, examsRepo)
	createLabUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, labExtractor)
	localExtractor := textextractioninfra.NewCommandExtractorWithOptions(textextractioninfra.CommandExtractorOptions{
		Timeout:           ocrConfig.Timeout,
		RequireUsableText: true,
	})
	textExtractor := textsvc.NewFallbackExtractor(localExtractor, fallback)
	authz := authorization.New(patientRepo, accessRepo)

	return &ExamsModule{
		Handler: handlers.NewExams(svc, createLabUC, storage, textExtractor, authz),
	}
}
