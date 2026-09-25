// internal/api/handlers/patient_access_test.go
package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type accessTestPatients struct {
	patientprofile.Repository
}

func (r accessTestPatients) FindByID(_ context.Context, id uuid.UUID) (*patient.Patient, error) {
	return &patient.Patient{ID: id}, nil
}

type accessTestGrants struct {
	repository.PatientAccessRepo
	t         *testing.T
	patientID uuid.UUID
	actorID   uuid.UUID
	called    bool
}

func (r *accessTestGrants) HasActiveAccess(_ context.Context, patientID, actorID uuid.UUID) (bool, error) {
	r.called = true
	if patientID != r.patientID || actorID != r.actorID {
		r.t.Fatal("access checked for the wrong patient or user")
	}
	return false, nil
}

func TestPatientDocumentsDenyAccessBeforeReadingOrProcessing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/labs"},
		{http.MethodGet, "/labs?full=true"},
		{http.MethodPost, "/labs"},
		{http.MethodGet, "/exames"},
		{http.MethodGet, "/exames/document-texts"},
		{http.MethodPost, "/exames"},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			actor := &accountdomain.User{ID: uuid.New(), AccountType: accountdomain.AccountTypeProfessional}
			patientID := uuid.New()
			grants := &accessTestGrants{t: t, patientID: patientID, actorID: actor.ID}
			authz := authorization.New(accessTestPatients{}, grants)
			// All data services are nil: reaching them after denial fails the test.
			labs := NewLabs(nil, nil, nil, authz)
			exams := NewExams(nil, nil, nil, nil, authz)
			router := gin.New()
			router.Use(func(c *gin.Context) { helpers.SetCurrentUser(c, actor) })
			router.GET("/patients/:id/labs", labs.ListLabs)
			router.POST("/patients/:id/labs", labs.UploadAndProcessLabs)
			router.GET("/patients/:id/exames", exams.ListExamDocuments)
			router.GET("/patients/:id/exames/document-texts", exams.ListExamDocumentTexts)
			router.POST("/patients/:id/exames", exams.UploadExamDocument)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(tc.method, "/patients/"+patientID.String()+tc.path, nil))
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
			}
			if !grants.called {
				t.Fatal("patient grant was not checked")
			}
		})
	}
}
