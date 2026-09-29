// internal/features/patient/exam/laboratory/http/handler.go
package laboratoryhttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
)

type labService interface {
	List(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]laboratory.LabReportSummaryOutput, error)
	ListFull(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*laboratory.LabReportOutput, error)
	FindByID(ctx context.Context, reportID uuid.UUID) (*laboratory.LabReportOutput, error)
}

type Handler struct {
	svc           labService
	accessChecker patientaccess.Checker
}

type listLabReportsInput struct {
	PatientID uuid.UUID `path:"patientId" format:"uuid"`
	Expand    string    `query:"expand" enum:"full"`
	Include   string    `query:"include"`
	Limit     int       `query:"limit" default:"100" minimum:"1" maximum:"100"`
	Offset    int       `query:"offset" default:"0" minimum:"0"`
}

type labReportInput struct {
	LabReportID uuid.UUID `path:"labReportId" format:"uuid"`
}

type listLabReportsOutput struct {
	Body any
}

type labReportOutput struct {
	Body laboratory.LabReportOutput
}

func NewHandler(svc labService, accessChecker patientaccess.Checker) *Handler {
	return &Handler{svc: svc, accessChecker: accessChecker}
}

// RegisterHumaRoutes registers the laboratory-report operations.
func (h *Handler) RegisterHumaRoutes(registered huma.API, security []map[string][]string) {
	huma.Register(registered, huma.Operation{
		OperationID: "listPatientLabReports",
		Method:      http.MethodGet,
		Path:        "/patients/{patientId}/lab-reports",
		Summary:     "Listar laudos laboratoriais",
		Tags:        []string{"Lab reports"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden},
		Security:    security,
	}, h.listLabReports)

	huma.Register(registered, huma.Operation{
		OperationID: "getLabReport",
		Method:      http.MethodGet,
		Path:        "/lab-reports/{labReportId}",
		Summary:     "Obter laudo laboratorial",
		Tags:        []string{"Lab reports"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
		Security:    security,
	}, h.getLabReport)
}

func (h *Handler) listLabReports(ctx context.Context, input *listLabReportsInput) (*listLabReportsOutput, error) {
	currentUser, err := humaCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.accessChecker.RequireAccess(ctx, currentUser.ID, input.PatientID); err != nil {
		return nil, toHumaError(err)
	}
	if shouldReturnFullLabsFor(input.Expand, input.Include) {
		list, err := h.svc.ListFull(ctx, input.PatientID, input.Limit, input.Offset)
		if err != nil {
			return nil, toHumaError(err)
		}
		return &listLabReportsOutput{Body: list}, nil
	}
	list, err := h.svc.List(ctx, input.PatientID, input.Limit, input.Offset)
	if err != nil {
		return nil, toHumaError(err)
	}
	return &listLabReportsOutput{Body: list}, nil
}

func (h *Handler) getLabReport(ctx context.Context, input *labReportInput) (*labReportOutput, error) {
	currentUser, err := humaCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	report, err := h.svc.FindByID(ctx, input.LabReportID)
	if err != nil {
		return nil, toHumaError(err)
	}
	if report == nil {
		return nil, huma.Error404NotFound("laudo não encontrado")
	}
	if err := h.accessChecker.RequireAccess(ctx, currentUser.ID, report.PatientID); err != nil {
		return nil, toHumaError(err)
	}
	return &labReportOutput{Body: *report}, nil
}

func humaCurrentUser(ctx context.Context) (*accountdomain.User, error) {
	currentUser, ok := helpers.GetCurrentUserFromContext(ctx)
	if !ok {
		return nil, huma.Error403Forbidden("conta registrada necessária")
	}
	return currentUser, nil
}

func toHumaError(err error) error {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) && appErr != nil {
		return huma.NewError(presenter.StatusFromCode(appErr.Kind), appErr.Message)
	}
	return huma.Error500InternalServerError("erro inesperado")
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

func (h *Handler) GetLabReport(c *gin.Context) {
	currentUser := helpers.MustGetCurrentUser(c)
	reportID, err := uuid.Parse(c.Param("labReportId"))
	if err != nil {
		presenter.ErrorResponder(c, apperr.Validation("lab_report_id inválido", apperr.Violation{Field: "lab_report_id", Reason: "invalid"}))
		return
	}
	report, err := h.svc.FindByID(c.Request.Context(), reportID)
	if err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	if report == nil {
		presenter.ErrorResponder(c, apperr.NotFound("laudo nao encontrado"))
		return
	}
	if err := h.accessChecker.RequireAccess(c.Request.Context(), currentUser.ID, report.PatientID); err != nil {
		presenter.ErrorResponder(c, err)
		return
	}
	c.JSON(http.StatusOK, report)
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
	return shouldReturnFullLabsFor(c.Query("expand"), c.Query("include"))
}

func shouldReturnFullLabsFor(expand, include string) bool {
	if strings.EqualFold(strings.TrimSpace(expand), "full") {
		return true
	}
	for _, raw := range strings.Split(include, ",") {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "full", "results", "test_results":
			return true
		}
	}
	return false
}
