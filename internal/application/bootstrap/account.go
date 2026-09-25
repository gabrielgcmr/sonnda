// internal/application/bootstrap/account.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/features/account"
	accounthttp "github.com/gabrielgcmr/sonnda/internal/features/account/http"
	accountpostgres "github.com/gabrielgcmr/sonnda/internal/features/account/postgres"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type AccountModule struct {
	Handler    *accounthttp.Handler
	Middleware *accounthttp.Middleware
}

func NewAccountModule(db *postgress.Client) *AccountModule {
	userRepo := accountpostgres.New(db.Pool())
	patientAccessRepo := repo.NewPatientAccessRepository(db)

	service := account.NewService(userRepo, patientAccessRepo)
	onboarding := account.NewOnboarding(userRepo, service)

	return &AccountModule{
		Handler:    accounthttp.NewHandler(onboarding, service),
		Middleware: accounthttp.NewMiddleware(userRepo),
	}
}
