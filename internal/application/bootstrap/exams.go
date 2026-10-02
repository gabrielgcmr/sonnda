// internal/application/bootstrap/exams.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/application/usecase/labdocumentconfirmation"
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	documents "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/extraction"
	documenthttp "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/http"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/labextraction"
	documentpostgres "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/postgres"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	accesspostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/access/postgres"
	labpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/postgres"
	patientpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/postgres"
	pginfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres"
	textinfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/textextraction"
)

type ExamsModule struct {
	Handler                        *documenthttp.ExamsHandler
	StandaloneLabExtractionHandler *documenthttp.StandaloneLabExtractionHandler
}

func NewExamsModule(db *pginfra.Client, lab labextraction.LabReportTextExtractor, storage documentprocessing.FileStorageService, config config.OCRConfig) *ExamsModule {
	patients := patientpostgres.NewRepository(db)
	labs := labpostgres.NewLabsRepository(db)
	service := documents.New(patients, documentpostgres.NewDocumentRepository(db))
	repo := documentpostgres.NewDraftRepository(db, labs)
	reader := textinfra.NewCommandExtractorWithOptions(textinfra.CommandExtractorOptions{Timeout: config.Timeout, RequireUsableText: true})
	extractor := extraction.New(reader, lab)
	drafts := documents.NewDrafts(repo, extractor, storage)
	confirmer := labdocumentconfirmation.New(service, drafts, repo, labs)
	access := patientaccess.NewChecker(patients, accesspostgres.NewRepository(db))
	return &ExamsModule{Handler: documenthttp.NewExams(service, drafts, confirmer, storage, access), StandaloneLabExtractionHandler: documenthttp.NewStandaloneLabExtraction(extractor)}
}
