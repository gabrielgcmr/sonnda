package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/api/handlers"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type ExamsModule struct {
	Handler *handlers.ExamsHandler
}

func NewExamsModule(
	dbClient *postgress.Client,
	storage domainstorage.FileStorageService,
) *ExamsModule {
	patientRepo := repo.NewPatientRepository(dbClient)
	accessRepo := repo.NewPatientAccessRepository(dbClient)
	profRepo := repo.NewProfessionalRepository(dbClient)
	examsRepo := repo.NewExamsRepository(dbClient)

	svc := examsvc.New(patientRepo, examsRepo)
	authz := authorization.New(patientRepo, accessRepo, profRepo)

	return &ExamsModule{
		Handler: handlers.NewExams(svc, storage, authz),
	}
}
