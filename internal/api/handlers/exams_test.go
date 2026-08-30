package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeExamsService struct {
	listByPatientCalled bool
	patientID           uuid.UUID
	limit               int
	offset              int
}

func (f *fakeExamsService) Create(ctx context.Context, input examsvc.CreateExamDocumentInput) (*examsvc.ExamDocumentOutput, error) {
	return nil, nil
}

func (f *fakeExamsService) FindByID(ctx context.Context, id uuid.UUID) (*examsvc.ExamDocumentOutput, error) {
	return nil, nil
}

func (f *fakeExamsService) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]examsvc.ExamDocumentOutput, error) {
	f.listByPatientCalled = true
	f.patientID = patientID
	f.limit = limit
	f.offset = offset
	return []examsvc.ExamDocumentOutput{}, nil
}

func (f *fakeExamsService) RouteDocument(ctx context.Context, input examsvc.RouteExamDocumentInput) (*examsvc.ExamDocumentOutput, error) {
	return nil, nil
}

func TestListExamDocuments_UsesServiceWithDefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeExamsService{}
	h := NewExams(svc, nil, allowAllAuthorizer{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/patients/:id/exames", h.ListExamDocuments)

	id := uuid.Must(uuid.NewV7())
	req := httptest.NewRequest(http.MethodGet, "/patients/"+id.String()+"/exames", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if !svc.listByPatientCalled {
		t.Fatal("expected ListByPatient to be called")
	}
	if svc.patientID != id {
		t.Fatalf("expected patientID %s, got %s", id, svc.patientID)
	}
	if svc.limit != 100 {
		t.Fatalf("expected default limit 100, got %d", svc.limit)
	}
	if svc.offset != 0 {
		t.Fatalf("expected default offset 0, got %d", svc.offset)
	}
}

func TestListExamDocuments_UsesQueryPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeExamsService{}
	h := NewExams(svc, nil, allowAllAuthorizer{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/patients/:id/exames", h.ListExamDocuments)

	id := uuid.Must(uuid.NewV7())
	req := httptest.NewRequest(http.MethodGet, "/patients/"+id.String()+"/exames?limit=25&offset=50", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if svc.limit != 25 {
		t.Fatalf("expected limit 25, got %d", svc.limit)
	}
	if svc.offset != 50 {
		t.Fatalf("expected offset 50, got %d", svc.offset)
	}
}
