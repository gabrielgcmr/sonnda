// internal/application/bootstrap/exams.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/api/handlers"
	textsvc "github.com/gabrielgcmr/sonnda/internal/application/services/textextraction"
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	examsvc "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	processingpostgres "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/postgres"
	labsuc "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/processing"
	laboratoryprocessor "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/processing/laboratory"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	accesspostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/access/postgres"
	labpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/postgres"
	patientpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/postgres"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	textextractioninfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/textextraction"
)

type ExamsModule struct {
	Handler *handlers.ExamsHandler
}

func NewExamsModule(
	dbClient *postgress.Client,
	labTextExtractor labextraction.LabReportTextExtractor,
	storage domainstorage.FileStorageService,
	ocrConfig config.OCRConfig,
	fallback domaintext.Extractor,
) *ExamsModule {
	patientRepo := patientpostgres.NewRepository(dbClient)
	accessRepo := accesspostgres.NewRepository(dbClient)
	examsRepo := processingpostgres.NewDocumentRepository(dbClient)
	labsRepo := processingpostgres.NewRepository(dbClient, labpostgres.NewLabsRepository(dbClient))

	svc := examsvc.New(patientRepo, examsRepo)
	createLabUC := labsuc.NewCreateLabReportFromDocument(patientRepo, labsRepo, nil)
	localExtractor := textextractioninfra.NewCommandExtractorWithOptions(textextractioninfra.CommandExtractorOptions{
		Timeout:           ocrConfig.Timeout,
		RequireUsableText: true,
	})
	textExtractor := textsvc.NewFallbackExtractor(localExtractor, fallback)
	var processors []labsuc.Processor
	if labTextExtractor != nil {
		processors = append(processors, laboratoryprocessor.NewProcessor(labTextExtractor, createLabUC, svc))
	}
	processDocumentUC := labsuc.NewProcessStoredDocument(svc, textExtractor, processors...)
	accessChecker := patientaccess.NewChecker(patientRepo, accessRepo)

	return &ExamsModule{
		Handler: handlers.NewExams(svc, processDocumentUC, storage, accessChecker),
	}
}
