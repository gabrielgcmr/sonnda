// internal/application/bootstrap/user.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/api/handlers"
	professionalsvc "github.com/gabrielgcmr/sonnda/internal/application/services/professional"
	usersvc "github.com/gabrielgcmr/sonnda/internal/application/services/user"
	registrationuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/registration"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type UserModule struct {
	Handler *handlers.UserHandler
}

func NewUserModule(db *postgress.Client) *UserModule {
	userRepo := repo.New(db)
	profRepo := repo.NewProfessionalRepository(db)
	patientAccessRepo := repo.NewPatientAccessRepository(db)

	userSvc := usersvc.New(userRepo, patientAccessRepo)
	profSvc := professionalsvc.New(profRepo)
	regUC := registrationuc.New(userRepo, userSvc, profSvc)

	handler := handlers.NewUserHandler(regUC, userSvc)

	return &UserModule{
		Handler: handler,
	}
}
