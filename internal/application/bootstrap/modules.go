// internal/application/bootstrap/modules.go
// internal/application/bootstrap/modules.go
// File: internal/app/bootstrap/modules.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
)

type Modules struct {
	User    *UserModule
	Patient *PatientModule
	Labs    *LabsModule
	Exams   *ExamsModule
}

func NewModules(
	dbClient *postgress.Client,
	labExtractor labextraction.LabReportExtractor,
	storage domainstorage.FileStorageService,
	ocrConfig config.OCRConfig,
) *Modules {
	return &Modules{
		User:    NewUserModule(dbClient),
		Patient: NewPatientModule(dbClient),
		Labs:    NewLabsModule(dbClient, labExtractor, storage),
		Exams:   NewExamsModule(dbClient, labExtractor, storage, ocrConfig),
	}
}
