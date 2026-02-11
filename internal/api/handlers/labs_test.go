// internal/api/handlers/labs_test.go
package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/rbac"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeLabsService struct {
	listCalled     bool
	listFullCalled bool
}

type allowAllAuthorizer struct{}

func (a allowAllAuthorizer) Require(ctx context.Context, actor *user.User, action rbac.Action, patientID *uuid.UUID) error {
	return nil
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
	h := NewLabs(svc, nil, nil, allowAllAuthorizer{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
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
	h := NewLabs(svc, nil, nil, allowAllAuthorizer{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
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
	h := NewLabs(svc, nil, nil, allowAllAuthorizer{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
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
