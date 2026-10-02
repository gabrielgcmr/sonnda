// internal/application/bootstrap/modules.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	"github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/labextraction"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres"
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
	labTextExtractor labextraction.LabReportTextExtractor,
	storage documentprocessing.FileStorageService,
	ocrConfig config.OCRConfig,
) *Modules {
	return &Modules{
		Account:       NewAccountModule(dbClient),
		Patient:       NewPatientModule(dbClient),
		PatientAccess: NewPatientAccessModule(dbClient),
		Labs:          NewLabsModule(dbClient),
		Exams:         NewExamsModule(dbClient, labTextExtractor, storage, ocrConfig),
	}
}
