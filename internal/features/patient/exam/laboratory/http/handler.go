// internal/features/patient/exam/laboratory/http/handler.go
package laboratoryhttp

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type labService interface {
	List(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]laboratory.LabReportSummaryOutput, error)
	ListFull(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*laboratory.LabReportOutput, error)
}

type Handler struct {
	svc           labService
	accessChecker patientaccess.Checker
}

func NewHandler(svc labService, accessChecker patientaccess.Checker) *Handler {
	return &Handler{svc: svc, accessChecker: accessChecker}
}

func (h *Handler) ListLabs(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	patientID, err := uuid.Parse(c.Param("patientId"))
	if err != nil {
		presenter.ErrorResponder(c, apperr.Validation("patient_id inválido", apperr.Violation{Field: "patient_id", Reason: "invalid"}))
		return
	}
	if err := h.accessChecker.RequireAccess(c.Request.Context(), currentUser.ID, patientID); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}

	limit, offset, ok := parsePagination(c)
	if !ok {
		return
	}

	if shouldReturnFullLabs(c) {
		list, err := h.svc.ListFull(c.Request.Context(), patientID, limit, offset)
		if err != nil {
			presenter.ErrorResponder(c, err)
			return
		}
		c.JSON(http.StatusOK, list)
		return
	}

	list, err := h.svc.List(c.Request.Context(), patientID, limit, offset)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func parsePagination(c *gin.Context) (limit, offset int, ok bool) {
	limit = 100
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			presenter.ErrorResponder(c, apperr.Validation("limit deve ser > 0", apperr.Violation{Field: "limit", Reason: "invalid"}))
			return 0, 0, false
		}
		limit = parsed
	}
	if raw := c.Query("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			presenter.ErrorResponder(c, apperr.Validation("offset deve ser >= 0", apperr.Violation{Field: "offset", Reason: "invalid"}))
			return 0, 0, false
		}
		offset = parsed
	}
	return limit, offset, true
}

func shouldReturnFullLabs(c *gin.Context) bool {
	if strings.EqualFold(strings.TrimSpace(c.Query("expand")), "full") {
		return true
	}
	for _, raw := range strings.Split(c.Query("include"), ",") {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "full", "results", "test_results":
			return true
		}
	}
	return false
}
