// internal/api/handlers/labs_handler_test.go
package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeLabsService struct {
	listCalled     bool
	listFullCalled bool
}

func (f *fakeLabsService) List(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]labsvc.LabReportSummaryOutput, error) {
	f.listCalled = true
	return []labsvc.LabReportSummaryOutput{}, nil
}

func (f *fakeLabsService) ListFull(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*labsvc.LabReportOutput, error) {
	f.listFullCalled = true
	return []*labsvc.LabReportOutput{}, nil
}

func TestListLabs_DefaultUsesSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeLabsService{}
	h := NewLabsHandler(svc, nil, nil)

	r := gin.New()
	r.GET("/patients/:id/labs", h.ListLabs)

	id := uuid.Must(uuid.NewV7()).String()
	req := httptest.NewRequest(http.MethodGet, "/patients/"+id+"/labs", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if !svc.listCalled {
		t.Fatal("expected List to be called")
	}
	if svc.listFullCalled {
		t.Fatal("did not expect ListFull to be called")
	}
}

func TestListLabs_ExpandFullUsesFull(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeLabsService{}
	h := NewLabsHandler(svc, nil, nil)

	r := gin.New()
	r.GET("/patients/:id/labs", h.ListLabs)

	id := uuid.Must(uuid.NewV7()).String()
	req := httptest.NewRequest(http.MethodGet, "/patients/"+id+"/labs?expand=full", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if !svc.listFullCalled {
		t.Fatal("expected ListFull to be called")
	}
	if svc.listCalled {
		t.Fatal("did not expect List to be called")
	}
}

func TestListLabs_IncludeResultsUsesFull(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeLabsService{}
	h := NewLabsHandler(svc, nil, nil)

	r := gin.New()
	r.GET("/patients/:id/labs", h.ListLabs)

	id := uuid.Must(uuid.NewV7()).String()
	req := httptest.NewRequest(http.MethodGet, "/patients/"+id+"/labs?include=results", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if !svc.listFullCalled {
		t.Fatal("expected ListFull to be called")
	}
	if svc.listCalled {
		t.Fatal("did not expect List to be called")
	}
}
