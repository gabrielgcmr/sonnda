// internal/application/bootstrap/patient.go
package bootstrap

import (
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	accesspostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/access/postgres"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	profilehttp "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/http"
	patientpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/postgres"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
)

type PatientModule struct {
	Service patientprofile.Service
	Handler *profilehttp.Handler
}

func NewPatientModule(db *postgress.Client) *PatientModule {
	patientRepo := patientpostgres.NewRepository(db)
	accessRepo := accesspostgres.NewRepository(db)

	accessChecker := patientaccess.NewChecker(patientRepo, accessRepo)
	svc := patientprofile.New(patientRepo, accessRepo, accessChecker)

	return &PatientModule{
		Service: svc,
		Handler: profilehttp.NewHandler(svc),
	}
}
