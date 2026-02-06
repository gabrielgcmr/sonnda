// internal/application/bootstrap/patient.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/api/handlers"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	patientsvc "github.com/gabrielgcmr/sonnda/internal/application/services/patient"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type PatientModule struct {
	Service patientsvc.Service
	Handler *handlers.PatientHandler
}

func NewPatientModule(db *postgress.Client) *PatientModule {
	patientRepo := repo.NewPatientRepository(db)
	accessRepo := repo.NewPatientAccessRepository(db)
	profRepo := repo.NewProfessionalRepository(db)

	authz := authorization.New(patientRepo, accessRepo, profRepo)
	svc := patientsvc.New(patientRepo, accessRepo, authz)

	return &PatientModule{
		Service: svc,
		Handler: handlers.NewPatientHandler(svc),
	}
}
