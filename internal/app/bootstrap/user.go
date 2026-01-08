package bootstrap

import (
	"sonnda-api/internal/adapters/inbound/http/api/handlers/user"
	"sonnda-api/internal/adapters/inbound/http/middleware"
	usersvc "sonnda-api/internal/app/services/user"

	repo "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	"sonnda-api/internal/domain/ports/integration"
)

type UserModule struct {
	Handler                *user.UserHandler
	RegistrationMiddleware *middleware.RegistrationMiddleware
}

func NewUserModule(db *db.Client, identityService integration.IdentityService) *UserModule {
	userRepo := repo.New(db)
	profRepo := repo.NewProfessionalRepository(db)
	patientRepo := repo.NewPatientRepository(db)
	accessRepo := repo.NewPatientAccessRepository(db)

	svc := usersvc.New(userRepo, profRepo, identityService)
	handler := user.NewUserHandler(svc, accessRepo)
	regMiddleware := middleware.NewRegistrationMiddleware(userRepo, patientRepo)

	return &UserModule{
		Handler:                handler,
		RegistrationMiddleware: regMiddleware,
	}
}
