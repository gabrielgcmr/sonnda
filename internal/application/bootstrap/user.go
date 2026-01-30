// internal/app/bootstrap/user.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/api/handlers/user"
	"github.com/gabrielgcmr/sonnda/internal/api/middleware"
	professionalsvc "github.com/gabrielgcmr/sonnda/internal/application/services/professional"
	usersvc "github.com/gabrielgcmr/sonnda/internal/application/services/user"
	registrationuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/registration"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type UserModule struct {
	Handler                *user.Handler
	RegistrationMiddleware *middleware.RegistrationMiddleware
}

func NewUserModule(db *postgress.Client) *UserModule {
	userRepo := repo.New(db)
	profRepo := repo.NewProfessionalRepository(db)
	patientAccessRepo := repo.NewPatientAccessRepository(db)

	userSvc := usersvc.New(userRepo, patientAccessRepo)
	profSvc := professionalsvc.New(profRepo)
	regUC := registrationuc.New(userRepo, userSvc, profSvc)

	handler := user.NewHandler(regUC, userSvc)
	regMiddleware := middleware.NewRegistrationMiddleware(userRepo)

	return &UserModule{
		Handler:                handler,
		RegistrationMiddleware: regMiddleware,
	}
}
