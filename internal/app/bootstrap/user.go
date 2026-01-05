package bootstrap

import (
	"sonnda-api/internal/app/interfaces/external"
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/http/api/handlers/user"
	"sonnda-api/internal/http/middleware"

	"sonnda-api/internal/infrastructure/persistence/repository/db"
	patientrepo "sonnda-api/internal/infrastructure/persistence/repository/patient"
	userrepo "sonnda-api/internal/infrastructure/persistence/repository/user"
)

type UserModule struct {
	Handler                *user.UserHandler
	RegistrationMiddleware *middleware.RegistrationMiddleware
}

func NewUserModule(db *db.Client, identityService external.IdentityService) *UserModule {
	userRepo := userrepo.New(db)
	profRepo := userrepo.NewProfessionalRepository(db)
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
