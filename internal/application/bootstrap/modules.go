// internal/application/bootstrap/modules.go
// internal/application/bootstrap/modules.go
// File: internal/app/bootstrap/modules.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/config"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
)

type Modules struct {
	Account       *UserModule
	Patient       *PatientModule
	MedicalRecord *MedicalRecordModule
}

type MedicalRecordModule struct {
	Labs  *LabsModule
	Exams *ExamsModule
}

func NewModules(
	dbClient *postgress.Client,
	labExtractor labextraction.LabReportExtractor,
	storage domainstorage.FileStorageService,
	ocrConfig config.OCRConfig,
	fallback domaintext.Extractor,
) *Modules {
	return &Modules{
		Account: NewUserModule(dbClient),
		Patient: NewPatientModule(dbClient),
		MedicalRecord: &MedicalRecordModule{
			Labs:  NewLabsModule(dbClient, labExtractor, storage),
			Exams: NewExamsModule(dbClient, labExtractor, storage, ocrConfig, fallback),
		},
	}
}
