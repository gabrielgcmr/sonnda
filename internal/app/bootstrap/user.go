// internal/app/bootstrap/user.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/handlers/user"
	"github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/api/middleware"
	sharedregistration "github.com/gabrielgcmr/sonnda/internal/adapters/inbound/http/shared/register"
	professionalsvc "github.com/gabrielgcmr/sonnda/internal/app/services/professional"
	usersvc "github.com/gabrielgcmr/sonnda/internal/app/services/user"
	registrationuc "github.com/gabrielgcmr/sonnda/internal/app/usecase/registration"

	postgress "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres"
	repo "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres/repository"
)

type UserModule struct {
	Handler                *user.Handler
	RegistrationCore       *sharedregistration.Core
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
	regCore := sharedregistration.NewCore(userRepo)
	regMiddleware := middleware.NewRegistrationMiddleware(regCore)

	return &UserModule{
		Handler:                handler,
		RegistrationCore:       regCore,
		RegistrationMiddleware: regMiddleware,
	}
}
