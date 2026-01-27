// internal/app/bootstrap/patient.go
// File: internal/app/bootstrap/patient.go
package bootstrap

import (
	patienthandler "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/handlers/patient"
	postgress "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres"
	repo "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres/repository"
	authorization "github.com/gabrielgcmr/sonnda/internal/app/services/authorization"
	patientsvc "github.com/gabrielgcmr/sonnda/internal/app/services/patient"
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
