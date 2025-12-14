package handlers

import (
	"log/slog"
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
		handleUserError(
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
		handleUserError(
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
		handleUserError(
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

func handleUserError(
	c *gin.Context,
	log *slog.Logger,
	status int,
	code string,
	err error,
) {
	if err != nil {
		log.Error("handler_error", "status", status, "error", code, "err", err)
		c.JSON(status, gin.H{
			"error":   code,
			"message": err.Error(),
		})
		return
	}

	// Sem err: casos como unauthorized podem ser Warn/Info (aqui usei Warn)
	log.Warn("handler_error", "status", status, "error", code)

	c.JSON(status, gin.H{"error": code})
}
