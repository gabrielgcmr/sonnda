// internal/features/patient/exam/laboratory/http/handler_test.go
package laboratoryhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
)

type fakeLabService struct {
	listCalled     bool
	listFullCalled bool
}

func (f *fakeLabService) List(context.Context, uuid.UUID, int, int) ([]laboratory.LabReportSummaryOutput, error) {
	f.listCalled = true
	return []laboratory.LabReportSummaryOutput{}, nil
}

func (f *fakeLabService) ListFull(context.Context, uuid.UUID, int, int) ([]*laboratory.LabReportOutput, error) {
	f.listFullCalled = true
	return []*laboratory.LabReportOutput{}, nil
}

type allowAllAccess struct{}

func (allowAllAccess) RequireAccess(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func TestListLabsReturnsSummaryByDefault(t *testing.T) {
	assertListMode(t, "", false)
}

func TestListLabsCanReturnFullResults(t *testing.T) {
	assertListMode(t, "?include=results", true)
	assertListMode(t, "?expand=full", true)
}

func assertListMode(t *testing.T, query string, wantFull bool) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	svc := &fakeLabService{}
	handler := NewHandler(svc, allowAllAccess{})
	router := gin.New()
	router.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &accountdomain.User{ID: uuid.New(), AccountType: accountdomain.AccountTypeBasicCare})
		c.Next()
	})
	router.GET("/patients/:patientId/labs", handler.ListLabs)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/patients/"+uuid.NewString()+"/labs"+query, nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if svc.listFullCalled != wantFull || svc.listCalled == wantFull {
		t.Fatalf("summary called=%v, full called=%v; want full=%v", svc.listCalled, svc.listFullCalled, wantFull)
	}
}
