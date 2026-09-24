// internal/application/bootstrap/account.go
package bootstrap

import (
	"github.com/gabrielgcmr/sonnda/internal/features/account"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"
)

type AccountModule struct {
	Middleware *account.Middleware
}

func NewAccountModule(db *postgress.Client) *AccountModule {
	return &AccountModule{
		Middleware: account.NewMiddleware(repo.New(db)),
	}
}
