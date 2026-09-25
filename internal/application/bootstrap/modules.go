// internal/application/bootstrap/modules.go
// internal/application/bootstrap/modules.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
)

type Modules struct {
	Account       *AccountModule
	Patient       *PatientModule
	PatientAccess *PatientAccessModule
	Labs          *LabsModule
	Exams         *ExamsModule
}

func NewModules(
	dbClient *postgress.Client,
	labExtractor labextraction.LabReportExtractor,
	storage domainstorage.FileStorageService,
	ocrConfig config.OCRConfig,
	fallback domaintext.Extractor,
) *Modules {
	return &Modules{
		Account:       NewAccountModule(dbClient),
		Patient:       NewPatientModule(dbClient),
		PatientAccess: NewPatientAccessModule(dbClient),
		Labs:          NewLabsModule(dbClient, labExtractor, storage),
		Exams:         NewExamsModule(dbClient, labExtractor, storage, ocrConfig, fallback),
	}
}
