package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/core/usecases/user"

	applog "sonnda-api/internal/logger"
)

type UserHandler struct {
	createUserFromIdentityUC *user.CreateUserFromIdentity
}

func NewUserHandler(
	createUserFromIdentityUC *user.CreateUserFromIdentity,
) *UserHandler {
	return &UserHandler{
		createUserFromIdentityUC: createUserFromIdentityUC,
	}
}

func (h *UserHandler) CreateUserFromIdentity(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	log.Info("creating_user_from_identity")

	identity, ok := middleware.Identity(c)
	if !ok {
		RespondError(
			c,
			log,
			http.StatusUnauthorized,
			"missing_identity",
			nil,
		)
		return
	}

	u, err := h.createUserFromIdentityUC.Execute(c.Request.Context(), identity)
	if err != nil {
		RespondError(
			c,
			log,
			http.StatusInternalServerError,
			"could_not_register_user",
			err,
		)
		return
	}

	log.Info("user_created_from_identity")
	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	log := applog.FromContext(c.Request.Context())
	user, ok := middleware.CurrentUser(c)
	if !ok {
		RespondError(
			c,
			log,
			http.StatusUnauthorized,
			"missing_user",
			nil,
		)
		return
	}

	c.JSON(http.StatusOK, user)
}
