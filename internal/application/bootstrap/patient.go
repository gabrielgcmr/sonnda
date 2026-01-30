// internal/application/bootstrap/patient.go
// File: internal/app/bootstrap/patient.go
package bootstrap

import (
	patienthandler "github.com/gabrielgcmr/sonnda/internal/api/handlers/patient"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	patientsvc "github.com/gabrielgcmr/sonnda/internal/application/services/patient"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type PatientModule struct {
	Service patientsvc.Service
	Handler *patienthandler.PatientHandler
}

func NewPatientModule(db *postgress.Client) *PatientModule {
	patientRepo := repo.NewPatientRepository(db)
	accessRepo := repo.NewPatientAccessRepository(db)
	profRepo := repo.NewProfessionalRepository(db)

	authz := authorization.New(patientRepo, accessRepo, profRepo)
	svc := patientsvc.New(patientRepo, authz)

	return &PatientModule{
		Service: svc,
		Handler: patienthandler.NewPatientHandler(svc),
	}
}
