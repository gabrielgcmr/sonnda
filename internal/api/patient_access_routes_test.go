// internal/api/patient_access_routes_test.go
package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
	accessdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/access/domain"
	accesshttp "github.com/gabrielgcmr/sonnda/internal/features/patient/access/http"
	"github.com/google/uuid"
)

func TestPatientAccessRouteOmitsRelationshipMetadata(t *testing.T) {
	router := newAccountRouter(&accountUserRepository{})
	accountRequest(t, router, http.MethodPost, "/v1/me", accountPayload, http.StatusCreated)

	response := accountRequest(t, router, http.MethodGet, "/v1/me/patients?limit=150&offset=2", "", http.StatusOK)
	var body struct {
		Patients []map[string]any `json:"patients"`
		Total    int64            `json:"total"`
		Limit    int              `json:"limit"`
		Offset   int              `json:"offset"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Limit != 100 || body.Offset != 2 || body.Total != 1 || len(body.Patients) != 1 {
		t.Fatalf("unexpected patient list: %+v", body)
	}
	if _, exposed := body.Patients[0]["relation_type"]; exposed {
		t.Fatalf("relationship metadata must not be exposed: %+v", body.Patients[0])
	}
}

func TestPatientAccessRouteRequiresAuthentication(t *testing.T) {
	router := newAccountRouter(&accountUserRepository{})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/me/patients", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func newPatientAccessTestHandler() *accesshttp.Handler {
	return accesshttp.NewHandler(patientaccess.NewService(patientAccessRouteRepository{}))
}

type patientAccessRouteRepository struct {
	patientaccess.Repository
}

func (patientAccessRouteRepository) ListAccessiblePatientsByUser(context.Context, uuid.UUID, int, int) ([]patientaccess.AccessiblePatient, int64, error) {
	return []patientaccess.AccessiblePatient{{
		PatientID:    uuid.MustParse("019a1f08-29a2-7b47-929d-bdc50bb59919"),
		FullName:     "Paciente Teste",
		RelationType: string(accessdomain.RelationshipTypeProfessional),
	}}, 1, nil
}
