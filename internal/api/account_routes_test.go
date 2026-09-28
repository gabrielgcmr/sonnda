// internal/api/account_routes_test.go
package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/api"
	patientcreation "github.com/gabrielgcmr/sonnda/internal/application/usecase/patientcreation"
	"github.com/gabrielgcmr/sonnda/internal/domain/demographics"
	"github.com/gabrielgcmr/sonnda/internal/features/account"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	accounthttp "github.com/gabrielgcmr/sonnda/internal/features/account/http"
	authdomain "github.com/gabrielgcmr/sonnda/internal/features/auth/domain"
	authhttp "github.com/gabrielgcmr/sonnda/internal/features/auth/http"
	patienthttp "github.com/gabrielgcmr/sonnda/internal/features/patient/http"
	patientprofile "github.com/gabrielgcmr/sonnda/internal/features/patient/profile"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"
	profilehttp "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const accountPayload = `{"full_name":"Ana Silva","birth_date":"1990-01-02","cpf":"12345678901","phone":"11999999999"}`

type accountUserResponse struct {
	ID          uuid.UUID `json:"id"`
	AuthIssuer  string    `json:"auth_issuer"`
	AuthSubject string    `json:"auth_subject"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	AccountType string    `json:"account_type"`
	BirthDate   string    `json:"birth_date"`
	CPF         string    `json:"cpf"`
}

func TestAccountRoutesLifecycle(t *testing.T) {
	repo := &accountUserRepository{}
	router := newAccountRouter(repo)

	accountRequest(t, router, http.MethodGet, "/me", "", http.StatusForbidden)

	created := decodeAccountUser(t, accountRequest(t, router, http.MethodPost, "/me", accountPayload, http.StatusCreated))
	if created.ID == uuid.Nil || created.FullName != "Ana Silva" || created.Email != "ana@example.test" ||
		created.AuthIssuer != "test-issuer" || created.AuthSubject != "test-subject" ||
		created.AccountType != string(accountdomain.AccountTypeBasicCare) || created.BirthDate != "1990-01-02" {
		t.Fatalf("unexpected created profile: %+v", created)
	}

	if current := decodeAccountUser(t, accountRequest(t, router, http.MethodGet, "/me", "", http.StatusOK)); current != created {
		t.Fatalf("GET profile = %+v, want %+v", current, created)
	}

	accountRequest(t, router, http.MethodPost, "/me", accountPayload, http.StatusConflict)

	updated := decodeAccountUser(t, accountRequest(t, router, http.MethodPut, "/me", `{"full_name":"Ana Souza"}`, http.StatusOK))
	if updated.ID != created.ID || updated.FullName != "Ana Souza" || updated.Email != created.Email || updated.CPF != created.CPF {
		t.Fatalf("unexpected updated profile: %+v", updated)
	}

	response := accountRequest(t, router, http.MethodDelete, "/me", "", http.StatusNoContent)
	if response.Body.Len() != 0 || repo.profile != nil {
		t.Fatal("deletion must remove the profile and return an empty response")
	}
	accountRequest(t, router, http.MethodGet, "/me", "", http.StatusForbidden)
}

func TestAccountRoutesReturnHumaErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		body   string
		err    error
		status int
	}{
		{"invalid JSON", `{`, nil, http.StatusBadRequest},
		{"missing birth date", `{"full_name":"Ana"}`, nil, http.StatusUnprocessableEntity},
		{"invalid CPF", strings.Replace(accountPayload, "12345678901", "123", 1), nil, http.StatusUnprocessableEntity},
		{"repository failure", accountPayload, errors.New("private database details"), http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &accountUserRepository{lookupErr: tc.err}
			response := accountRequest(t, newAccountRouter(repo), http.MethodPost, "/me", tc.body, tc.status)
			if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/problem+json") || strings.Contains(response.Body.String(), "private database details") {
				t.Fatalf("unexpected Huma error response: %s", response.Body.String())
			}
		})
	}
}

func TestAccountRoutesRequireBearerAuthentication(t *testing.T) {
	router := newAccountRouter(&accountUserRepository{})
	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/me"},
		{http.MethodGet, "/me"},
		{http.MethodPut, "/me"},
		{http.MethodDelete, "/me"},
		{http.MethodPost, "/patients/" + uuid.NewString() + "/exam-documents"},
		{http.MethodGet, "/exam-documents/" + uuid.NewString()},
		{http.MethodGet, "/patients/" + uuid.NewString() + "/lab-reports"},
		{http.MethodGet, "/lab-reports/" + uuid.NewString()},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(route.method, route.path, nil))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s without authentication returned %d", route.method, route.path, response.Code)
		}
	}
}

func newAccountRouter(repo *accountUserRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	service := account.NewService(repo)
	onboarding := account.NewOnboarding(repo, service)
	authMiddleware := authhttp.NewMiddleware(func(context.Context, string) (*authdomain.Identity, error) {
		email := "ana@example.test"
		return &authdomain.Identity{Issuer: "test-issuer", Subject: "test-subject", Email: &email}, nil
	})
	router := gin.New()
	api.SetupRoutes(router, &api.APIDependencies{
		Auth:                   authMiddleware,
		Account:                accounthttp.NewMiddleware(repo),
		AccountHandler:         accounthttp.NewHandler(onboarding, service),
		PatientAccessHandler:   newPatientAccessTestHandler(),
		PatientCreationHandler: patienthttp.NewCreationHandler(patientCreationRouteStub{}),
		PatientHandler:         profilehttp.NewHandler(patientProfileRouteStub{}),
	})
	return router
}

func accountRequest(t *testing.T, router http.Handler, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer test-token")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != status {
		t.Fatalf("%s %s: status = %d, want %d; body = %s", method, path, response.Code, status, response.Body.String())
	}
	return response
}

func decodeAccountUser(t *testing.T, response *httptest.ResponseRecorder) accountUserResponse {
	t.Helper()
	var profile accountUserResponse
	if err := json.Unmarshal(response.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	return profile
}

type accountUserRepository struct {
	account.Repository
	profile   *accountdomain.User
	lookupErr error
}

func (r *accountUserRepository) FindByAuthIdentity(_ context.Context, issuer, subject string) (*accountdomain.User, error) {
	if r.lookupErr != nil {
		return nil, r.lookupErr
	}
	if r.profile == nil || r.profile.AuthIssuer != issuer || r.profile.AuthSubject != subject {
		return nil, nil
	}
	profile := *r.profile
	return &profile, nil
}

func (r *accountUserRepository) FindByID(_ context.Context, id uuid.UUID) (*accountdomain.User, error) {
	if r.profile == nil || r.profile.ID != id {
		return nil, nil
	}
	profile := *r.profile
	return &profile, nil
}

func (r *accountUserRepository) Create(_ context.Context, profile *accountdomain.User) error {
	copy := *profile
	r.profile = &copy
	return nil
}

func (r *accountUserRepository) Update(ctx context.Context, profile *accountdomain.User) error {
	return r.Create(ctx, profile)
}

func (r *accountUserRepository) Delete(_ context.Context, id uuid.UUID) error {
	if r.profile != nil && r.profile.ID == id {
		r.profile = nil
	}
	return nil
}

type patientCreationRouteStub struct{}

func (patientCreationRouteStub) Execute(_ context.Context, accountID uuid.UUID, _ patientcreation.Input) (*profiledomain.Patient, error) {
	return &profiledomain.Patient{ID: uuid.MustParse("019a1f08-29a2-7b47-929d-bdc50bb59919"), OwnerUserID: &accountID}, nil
}

type patientProfileRouteStub struct{}

func (patientProfileRouteStub) Get(_ context.Context, currentUser *accountdomain.User, id uuid.UUID) (*profiledomain.Patient, error) {
	return patientRoutePatient(currentUser.ID, id), nil
}

func (patientProfileRouteStub) Update(context.Context, *accountdomain.User, uuid.UUID, patientprofile.UpdateInput) (*profiledomain.Patient, error) {
	return nil, errors.New("unused")
}

func (patientProfileRouteStub) HardDelete(context.Context, *accountdomain.User, uuid.UUID) error {
	return errors.New("unused")
}

func (patientProfileRouteStub) ListMyPatients(_ context.Context, currentUser *accountdomain.User, _, _ int) ([]*profiledomain.Patient, error) {
	return []*profiledomain.Patient{patientRoutePatient(currentUser.ID, uuid.MustParse("019a1f08-29a2-7b47-929d-bdc50bb59919"))}, nil
}

func patientRoutePatient(ownerID, patientID uuid.UUID) *profiledomain.Patient {
	return &profiledomain.Patient{
		ID:          patientID,
		OwnerUserID: &ownerID,
		CPF:         "12345678901",
		FullName:    "Paciente Teste",
		BirthDate:   time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		Gender:      demographics.GenderFemale,
		Race:        demographics.RaceWhite,
		CreatedAt:   time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
}
