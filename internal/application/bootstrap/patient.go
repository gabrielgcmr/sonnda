// internal/application/bootstrap/patient.go
package bootstrap

import (
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
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

	authz := authorization.New(patientRepo, accessRepo)
	svc := patientprofile.New(patientRepo, accessRepo, authz)

	return &PatientModule{
		Service: svc,
		Handler: profilehttp.NewHandler(svc),
	}
}
