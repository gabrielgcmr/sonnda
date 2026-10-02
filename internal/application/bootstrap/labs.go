// internal/application/bootstrap/labs.go
package bootstrap

import (
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	accesspostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/access/postgres"
	labsvc "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	laboratoryhttp "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/http"
	labpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/postgres"
	patientpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/postgres"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres"
)

type LabsModule struct {
	LaboratoryHandler *laboratoryhttp.Handler
}

func NewLabsModule(dbClient *postgress.Client) *LabsModule {
	patientRepo := patientpostgres.NewRepository(dbClient)
	accessRepo := accesspostgres.NewRepository(dbClient)
	labsRepo := labpostgres.NewRepository(dbClient)

	svc := labsvc.New(patientRepo, labsRepo)
	accessChecker := patientaccess.NewChecker(patientRepo, accessRepo)
	return &LabsModule{
		LaboratoryHandler: laboratoryhttp.NewHandler(svc, accessChecker),
	}
}
