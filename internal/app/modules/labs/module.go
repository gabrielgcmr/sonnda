package labs

import (
	"sonnda-api/internal/app/usecases/labuc"
	"sonnda-api/internal/domain/ports/integrations"
	"sonnda-api/internal/domain/ports/repositories"
	labs "sonnda-api/internal/http/api/handlers/labs"
	"sonnda-api/internal/infrastructure/persistence/repository"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
)

type Module struct {
	Repo       repositories.LabRepository
	ExtractUC  *labuc.ExtractLabsUseCase
	ListUC     *labuc.ListLabsUseCase
	ListFullUC *labuc.ListFullLabsUseCase
	Handler    *labs.LabsHandler
}

func New(
	db *db.Client,
	patientRepo repositories.PatientRepository,
	docExtractor integrations.DocumentExtractor,
	storageService integrations.StorageService,
) *Module {
	repo := repository.NewLabsRepository(db)
	extractUC := labuc.NewExtractLabs(repo, docExtractor)
	listUC := labuc.NewListLabs(patientRepo, repo)
	listFullUC := labuc.NewListFullLabs(patientRepo, repo)
	handler := labs.NewLabsHandler(extractUC, listUC, listFullUC, storageService)

	return &Module{
		Repo:       repo,
		ExtractUC:  extractUC,
		ListUC:     listUC,
		ListFullUC: listFullUC,
		Handler:    handler,
	}
}
