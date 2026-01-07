// internal/app/bootstrap/patient.go
package bootstrap

import (
	patientsvc "sonnda-api/internal/app/services/patient"
	patienthandler "sonnda-api/internal/adapters/inbound/http/api/handlers/patient"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	patientrepo "sonnda-api/internal/adapters/outbound/persistence/repository/patient"
)

func NewPatientModule(db *db.Client) *patienthandler.PatientHandler {
	repo := patientrepo.NewPatientRepository(db)

	svc := patientsvc.New(repo, nil)

	return patienthandler.NewPatientHandler(svc)
}
