package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	helpers "github.com/gabrielgcmr/sonnda/internal/api/helpers"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeExamsService struct {
	listByPatientCalled bool
	listReportsCalled   bool
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

func (f *fakeExamsService) MarkFailed(ctx context.Context, input examsvc.MarkExamDocumentFailedInput) (*examsvc.ExamDocumentOutput, error) {
	return nil, nil
}

func (f *fakeExamsService) CreateReportFromText(ctx context.Context, input examsvc.CreateExamReportFromTextInput) (*examsvc.ExamReportOutput, error) {
	return nil, nil
}

func (f *fakeExamsService) ListReportsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]examsvc.ExamReportOutput, error) {
	f.listReportsCalled = true
	f.patientID = patientID
	f.limit = limit
	f.offset = offset
	return []examsvc.ExamReportOutput{}, nil
}

func TestListExamDocuments_UsesServiceWithDefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeExamsService{}
	h := NewExams(svc, nil, nil, nil, allowAllAuthorizer{})

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

func TestListExamReports_UsesService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeExamsService{}
	h := NewExams(svc, nil, nil, nil, allowAllAuthorizer{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		helpers.SetCurrentUser(c, &user.User{ID: uuid.Must(uuid.NewV7()), AccountType: user.AccountTypeBasicCare})
		c.Next()
	})
	r.GET("/patients/:id/exames/reports", h.ListExamReports)

	id := uuid.Must(uuid.NewV7())
	req := httptest.NewRequest(http.MethodGet, "/patients/"+id.String()+"/exames/reports", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if !svc.listReportsCalled {
		t.Fatal("expected ListReportsByPatient to be called")
	}
}

func TestListExamDocuments_UsesQueryPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeExamsService{}
	h := NewExams(svc, nil, nil, nil, allowAllAuthorizer{})

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

func TestBuildLabReportText_FormatsStructuredResults(t *testing.T) {
	reportDate := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)
	patientName := "Gabriel Cactus Moreno Reboucas"
	labName := "Laboratorio Exemplo"
	unit := "g/dL"
	value := "15,1"
	reference := "13,5 a 17,5"

	text := buildLabReportText(&labsvc.LabReportOutput{
		PatientName: &patientName,
		LabName:     &labName,
		ReportDate:  &reportDate,
		TestResults: []labsvc.TestResultOutput{
			{
				TestName: "HEMOGRAMA",
				Items: []labsvc.TestItemOutput{
					{
						ParameterName: "Hemoglobina",
						ResultValue:   &value,
						ResultUnit:    &unit,
						ReferenceText: &reference,
					},
				},
			},
		},
	})

	expectedParts := []string{
		"Exame laboratorial",
		"Paciente: Gabriel Cactus Moreno Reboucas",
		"Laboratorio: Laboratorio Exemplo",
		"Data do laudo: 31/08/2026",
		"HEMOGRAMA",
		"- Hemoglobina 15,1 g/dL (Referencia: 13,5 a 17,5)",
	}
	for _, part := range expectedParts {
		if !strings.Contains(text, part) {
			t.Fatalf("expected text to contain %q, got:\n%s", part, text)
		}
	}
}
