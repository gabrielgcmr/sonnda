// internal/app/bootstrap/user.go
package bootstrap

import (
	"sonnda-api/internal/adapters/inbound/http/api/handlers/user"
	"sonnda-api/internal/adapters/inbound/http/middleware"
	professionalsvc "sonnda-api/internal/app/services/professional"
	usersvc "sonnda-api/internal/app/services/user"
	registrationuc "sonnda-api/internal/app/usecase/registration"

	repo "sonnda-api/internal/adapters/outbound/persistence/repository"
	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	"sonnda-api/internal/domain/ports/integration"
)

type UserModule struct {
	Handler                *user.Handler
	RegistrationMiddleware *middleware.RegistrationMiddleware
}

func NewUserModule(db *db.Client, identityService integration.IdentityService) *UserModule {
	userRepo := repo.New(db)
	profRepo := repo.NewProfessionalRepository(db)
	patientRepo := repo.NewPatientRepository(db)
	patientAccessRepo := repo.NewPatientAccessRepository(db)

	userSvc := usersvc.New(userRepo, patientAccessRepo)
	profSvc := professionalsvc.New(profRepo)
	regUC := registrationuc.New(userRepo, userSvc, profSvc, identityService)

	handler := user.NewHandler(regUC, userSvc)
	regMiddleware := middleware.NewRegistrationMiddleware(userRepo, patientRepo)

	return &UserModule{
		Handler:                handler,
		RegistrationMiddleware: regMiddleware,
	}
}
