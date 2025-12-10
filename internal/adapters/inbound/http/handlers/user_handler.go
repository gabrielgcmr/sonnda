package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sonnda-api/internal/adapters/inbound/http/middleware"
	"sonnda-api/internal/core/usecases/user"
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
	identity, ok := middleware.Identity(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	u, err := h.createUserFromIdentityUC.Execute(c.Request.Context(), identity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "could_not_register_user",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, user)
}
