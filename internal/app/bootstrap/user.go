package bootstrap

import (
	usersvc "sonnda-api/internal/app/services/user"
	"sonnda-api/internal/infrastructure/persistence/repository/db"
)

func newUserModule(db *db.Client) *userhandler.UserHandler {
	repo := userrepo.NewUserRepository(db)

	svc := usersvc.New(repo, nil)
	return userhandler.NewUserHandler(svc)
}
