// internal/app/bootstrap/patient.go
package bootstrap

import (
	patienthandler "sonnda-api/internal/adapters/inbound/http/api/handlers/patient"
	patientrepo "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	patientsvc "sonnda-api/internal/app/services/patient"
)

func NewPatientModule(db *db.Client) *patienthandler.PatientHandler {
	repo := patientrepo.NewPatientRepository(db)

	svc := patientsvc.New(repo, nil)

	return patienthandler.NewPatientHandler(svc)
}
