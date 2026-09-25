// internal/features/patient/access/http/handler.go
package accesshttp

import (
	"net/http"
	"strconv"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service patientaccess.Service
}

func NewHandler(service patientaccess.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListForCurrentAccount(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	limit, offset := accessPagination(c)

	result, err := h.service.ListForAccount(c.Request.Context(), currentUser.ID, limit, offset)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	c.JSON(http.StatusOK, listPatientsResponse(result))
}

func accessPagination(c *gin.Context) (int, int) {
	limit, offset := 20, 0
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if raw := c.Query("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}
