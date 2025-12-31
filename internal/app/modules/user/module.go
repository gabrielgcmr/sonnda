package user

import (
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/domain/ports/integrations"
	"sonnda-api/internal/domain/ports/repositories"
	userhandler "sonnda-api/internal/http/api/handlers/user"
	"sonnda-api/internal/http/middleware"
	repository "sonnda-api/internal/infrastructure/persistence/repository"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
	userrepo "sonnda-api/internal/infrastructure/persistence/repository/user"
)

type Module struct {
	Handler                *userhandler.UserHandler
	RegistrationMiddleware *middleware.RegistrationMiddleware
}

func New(
	db *db.Client,
	patientRepo repositories.PatientRepository,
	accessRepo repositories.PatientAccessRepository,
	identityService integrations.IdentityService,
) *Module {
	// Repositórios
	userRepo := userrepo.New(db)
	profRepo := repository.NewProfessionalRepository(db)

	// Service
	svc := usersvc.New(userRepo, profRepo, identityService)

	// Handlers
	handler := userhandler.NewUserHandler(svc, accessRepo)

	// Middlewares
	regMiddleware := middleware.NewRegistrationMiddleware(userRepo, patientRepo)

	return &Module{
		Handler:                handler,
		RegistrationMiddleware: regMiddleware,
	}
}
