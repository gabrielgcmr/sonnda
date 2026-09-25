// internal/api/handlers/patient_access_test.go
package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	laboratoryhttp "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/http"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type allowAllAccessChecker struct{}

func (allowAllAccessChecker) RequireAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type accessTestPatients struct {
	patientprofile.Repository
}

func (r accessTestPatients) FindByID(_ context.Context, id uuid.UUID) (*profiledomain.Patient, error) {
	return &profiledomain.Patient{ID: id}, nil
}

type accessTestGrants struct {
	patientaccess.Repository
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
			accessChecker := patientaccess.NewChecker(accessTestPatients{}, grants)
			// All data services are nil: reaching them after denial fails the test.
			labs := NewLabs(nil, nil, accessChecker)
			laboratory := laboratoryhttp.NewHandler(nil, accessChecker)
			exams := NewExams(nil, nil, nil, accessChecker)
			router := gin.New()
			router.Use(func(c *gin.Context) { helpers.SetCurrentUser(c, actor) })
			router.GET("/patients/:patientId/labs", laboratory.ListLabs)
			router.POST("/patients/:patientId/labs", labs.UploadAndProcessLabs)
			router.GET("/patients/:patientId/exames", exams.ListExamDocuments)
			router.GET("/patients/:patientId/exames/document-texts", exams.ListExamDocumentTexts)
			router.POST("/patients/:patientId/exames", exams.UploadExamDocument)
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
