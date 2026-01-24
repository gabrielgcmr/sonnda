// File: internal/app/bootstrap/patient.go
package bootstrap

import (
	patienthandler "sonnda-api/internal/adapters/inbound/http/api/handlers/patient"
	repo "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	authorization "sonnda-api/internal/app/services/authorization"
	patientsvc "sonnda-api/internal/app/services/patient"
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
