// internal/app/bootstrap/patient.go
package bootstrap

import (
	patientsvc "sonnda-api/internal/app/services/patient"
	patienthandler "sonnda-api/internal/http/api/handlers/patient"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
	patientrepo "sonnda-api/internal/infrastructure/persistence/repository/patient"
)

func NewPatientModule(db *db.Client) *patienthandler.PatientHandler {
	repo := patientrepo.NewPatientRepository(db)

	svc := patientsvc.New(repo, nil)

	return patienthandler.NewPatientHandler(svc)
}
