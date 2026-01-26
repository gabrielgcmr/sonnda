// internal/app/bootstrap/patient.go
// File: internal/app/bootstrap/patient.go
package bootstrap

import (
	patienthandler "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/handlers/patient"
	repo "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/data/postgres/repository"
	"github.com/gabrielgcmr/sonnda/internal/adapters/outbound/data/postgres/repository/db"
	authorization "github.com/gabrielgcmr/sonnda/internal/app/services/authorization"
	patientsvc "github.com/gabrielgcmr/sonnda/internal/app/services/patient"
)

type PatientModule struct {
	Service patientsvc.Service
	Handler *patienthandler.PatientHandler
}

func NewPatientModule(db *db.Client) *PatientModule {
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
