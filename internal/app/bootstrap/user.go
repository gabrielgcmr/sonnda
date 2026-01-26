// internal/app/bootstrap/user.go
package bootstrap

import (
	"sonnda-api/internal/adapters/inbound/http/api/handlers/user"
	"sonnda-api/internal/adapters/inbound/http/api/middleware"
	sharedregistration "sonnda-api/internal/adapters/inbound/http/shared/registration"
	professionalsvc "sonnda-api/internal/app/services/professional"
	usersvc "sonnda-api/internal/app/services/user"
	registrationuc "sonnda-api/internal/app/usecase/registration"

	repo "sonnda-api/internal/adapters/outbound/data/postgres/repository"
	"sonnda-api/internal/adapters/outbound/data/postgres/repository/db"
	"sonnda-api/internal/domain/ports"
)

type UserModule struct {
	Handler                *user.Handler
	RegistrationCore       *sharedregistration.Core
	RegistrationMiddleware *middleware.RegistrationMiddleware
}

func NewUserModule(db *db.Client, identityService ports.IdentityService) *UserModule {
	userRepo := repo.New(db)
	profRepo := repo.NewProfessionalRepository(db)
	patientAccessRepo := repo.NewPatientAccessRepository(db)

	userSvc := usersvc.New(userRepo, patientAccessRepo)
	profSvc := professionalsvc.New(profRepo)
	regUC := registrationuc.New(userRepo, userSvc, profSvc, identityService)

	handler := user.NewHandler(regUC, userSvc)
	regCore := sharedregistration.NewCore(userRepo)
	regMiddleware := middleware.NewRegistrationMiddleware(regCore)

	return &UserModule{
		Handler:                handler,
		RegistrationCore:       regCore,
		RegistrationMiddleware: regMiddleware,
	}
}
