package bootstrap

import (
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/adapters/inbound/http/api/handlers/user"
	"sonnda-api/internal/adapters/inbound/http/middleware"

	"sonnda-api/internal/adapters/outbound/persistence/repository/db"
	patientrepo "sonnda-api/internal/adapters/outbound/persistence/repository/patient"
	professionalrepo "sonnda-api/internal/adapters/outbound/persistence/repository/professional"
	userrepo "sonnda-api/internal/adapters/outbound/persistence/repository/user"
	"sonnda-api/internal/domain/ports/integrations"
)

type UserModule struct {
	Handler                *user.UserHandler
	RegistrationMiddleware *middleware.RegistrationMiddleware
}

func NewUserModule(db *db.Client, identityService integrations.IdentityService) *UserModule {
	userRepo := userrepo.New(db)
	profRepo := professionalrepo.NewProfessionalRepository(db)
	patientRepo := patientrepo.NewPatientRepository(db)
	accessRepo := patientrepo.NewPatientAccessRepository(db)

	svc := usersvc.New(userRepo, profRepo, identityService)
	handler := user.NewUserHandler(svc, accessRepo)
	regMiddleware := middleware.NewRegistrationMiddleware(userRepo, patientRepo)

	return &UserModule{
		Handler:                handler,
		RegistrationMiddleware: regMiddleware,
	}
}
