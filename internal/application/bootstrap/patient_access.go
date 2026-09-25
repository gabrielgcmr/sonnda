// internal/application/bootstrap/patient_access.go
package bootstrap

import (
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	accesshttp "github.com/gabrielgcmr/sonnda/internal/features/patient/access/http"
	accesspostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/access/postgres"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
)

type PatientAccessModule struct {
	Handler *accesshttp.Handler
}

func NewPatientAccessModule(db *postgress.Client) *PatientAccessModule {
	repo := accesspostgres.NewRepository(db)
	service := patientaccess.NewService(repo)
	return &PatientAccessModule{Handler: accesshttp.NewHandler(service)}
}
