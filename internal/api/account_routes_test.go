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

	"github.com/gabrielgcmr/sonnda/internal/api"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/identity"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/features/account"
	accounthttp "github.com/gabrielgcmr/sonnda/internal/features/account/http"
	"github.com/gabrielgcmr/sonnda/internal/features/auth"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const accountPayload = `{"full_name":"Ana Silva","birth_date":"1990-01-02","cpf":"12345678901","phone":"11999999999"}`

func TestAccountRoutesLifecycle(t *testing.T) {
	for _, createPath := range []string{"/v1/me", "/v1/users"} {
		t.Run(createPath, func(t *testing.T) {
			repo := &accountUserRepository{}
			router := newAccountRouter(repo)

			response := accountRequest(t, router, http.MethodGet, "/v1/me", "", http.StatusForbidden)
			assertAccountProblem(t, response, apperr.PROFILE_NOT_FOUND)

			response = accountRequest(t, router, http.MethodPost, createPath, accountPayload, http.StatusCreated)
			created := decodeAccountUser(t, response)
			if created.ID == uuid.Nil || created.FullName != "Ana Silva" || created.Email != "ana@example.test" ||
				created.AuthIssuer != "test-issuer" || created.AuthSubject != "test-subject" ||
				created.AccountType != user.AccountTypeBasicCare || created.BirthDate.Format("2006-01-02") != "1990-01-02" {
				t.Fatalf("unexpected created profile: %+v", created)
			}

			response = accountRequest(t, router, http.MethodGet, "/v1/me", "", http.StatusOK)
			if current := decodeAccountUser(t, response); current != created {
				t.Fatalf("GET profile = %+v, want %+v", current, created)
			}

			response = accountRequest(t, router, http.MethodPost, createPath, accountPayload, http.StatusConflict)
			assertAccountProblem(t, response, apperr.RESOURCE_ALREADY_EXISTS)

			response = accountRequest(t, router, http.MethodPut, "/v1/me", `{"full_name":"Ana Souza"}`, http.StatusOK)
			updated := decodeAccountUser(t, response)
			if updated.ID != created.ID || updated.FullName != "Ana Souza" || updated.Email != created.Email || updated.CPF != created.CPF {
				t.Fatalf("unexpected updated profile: %+v", updated)
			}
			response = accountRequest(t, router, http.MethodGet, "/v1/me", "", http.StatusOK)
			if current := decodeAccountUser(t, response); current != updated {
				t.Fatalf("update was not persisted: %+v", current)
			}

			response = accountRequest(t, router, http.MethodGet, "/v1/me/patients?limit=150&offset=2", "", http.StatusOK)
			var patients account.MyPatientsOutput
			if err := json.Unmarshal(response.Body.Bytes(), &patients); err != nil {
				t.Fatal(err)
			}
			if patients.Limit != 100 || patients.Offset != 2 || patients.Total != 0 || patients.Patients == nil {
				t.Fatalf("unexpected patient list: %+v", patients)
			}

			response = accountRequest(t, router, http.MethodDelete, "/v1/me", "", http.StatusNoContent)
			if response.Body.Len() != 0 || repo.profile != nil {
				t.Fatal("deletion must remove the profile and return an empty response")
			}
			response = accountRequest(t, router, http.MethodGet, "/v1/me", "", http.StatusForbidden)
			assertAccountProblem(t, response, apperr.PROFILE_NOT_FOUND)
		})
	}
}

func TestAccountRoutesPreserveErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		body   string
		err    error
		status int
		code   apperr.ErrorKind
	}{
		{"invalid JSON", `{`, nil, http.StatusBadRequest, apperr.VALIDATION_FAILED},
		{"missing birth date", `{"full_name":"Ana"}`, nil, http.StatusBadRequest, apperr.VALIDATION_FAILED},
		{"invalid CPF", strings.Replace(accountPayload, "12345678901", "123", 1), nil, http.StatusBadRequest, apperr.VALIDATION_FAILED},
		{"repository failure", accountPayload, errors.New("private database details"), http.StatusInternalServerError, apperr.INTERNAL_ERROR},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &accountUserRepository{lookupErr: tc.err}
			response := accountRequest(t, newAccountRouter(repo), http.MethodPost, "/v1/me", tc.body, tc.status)
			assertAccountProblem(t, response, tc.code)
			if repo.profile != nil || strings.Contains(response.Body.String(), "private database details") {
				t.Fatal("failed registration must not persist a profile or expose the internal error")
			}
		})
	}

	router := newAccountRouter(&accountUserRepository{})
	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/v1/me"},
		{http.MethodPost, "/v1/users"},
		{http.MethodGet, "/v1/me"},
		{http.MethodPut, "/v1/me"},
		{http.MethodDelete, "/v1/me"},
		{http.MethodGet, "/v1/me/patients"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(route.method, route.path, nil))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s without authentication returned %d", route.method, route.path, response.Code)
		}
		assertAccountProblem(t, response, apperr.AUTH_REQUIRED)
	}
}

func newAccountRouter(repo *accountUserRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	service := account.NewService(repo, accountPatientAccessRepository{})
	onboarding := account.NewOnboarding(repo, service, nil)
	authMiddleware := auth.NewMiddleware(func(context.Context, string) (*identity.Identity, error) {
		email := "ana@example.test"
		return &identity.Identity{Issuer: "test-issuer", Subject: "test-subject", Email: &email}, nil
	})
	router := gin.New()
	api.SetupRoutes(router, &api.APIDependencies{
		Auth:           authMiddleware,
		Account:        accounthttp.NewMiddleware(repo),
		AccountHandler: accounthttp.NewHandler(onboarding, service),
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

func decodeAccountUser(t *testing.T, response *httptest.ResponseRecorder) user.User {
	t.Helper()
	var profile user.User
	if err := json.Unmarshal(response.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	return profile
}

func assertAccountProblem(t *testing.T, response *httptest.ResponseRecorder, code apperr.ErrorKind) {
	t.Helper()
	var problem presenter.Problem
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if problem.Code != code || problem.Status != response.Code {
		t.Fatalf("unexpected error contract: %+v", problem)
	}
	if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/problem+json") {
		t.Fatalf("unexpected error content type: %s", response.Header().Get("Content-Type"))
	}
}

type accountUserRepository struct {
	account.Repository
	profile   *user.User
	lookupErr error
}

func (r *accountUserRepository) FindByAuthIdentity(_ context.Context, issuer, subject string) (*user.User, error) {
	if r.lookupErr != nil {
		return nil, r.lookupErr
	}
	if r.profile == nil || r.profile.AuthIssuer != issuer || r.profile.AuthSubject != subject {
		return nil, nil
	}
	profile := *r.profile
	return &profile, nil
}

func (r *accountUserRepository) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	if r.profile == nil || r.profile.ID != id {
		return nil, nil
	}
	profile := *r.profile
	return &profile, nil
}

func (r *accountUserRepository) Create(_ context.Context, profile *user.User) error {
	copy := *profile
	r.profile = &copy
	return nil
}

func (r *accountUserRepository) Update(ctx context.Context, profile *user.User) error {
	return r.Create(ctx, profile)
}

func (r *accountUserRepository) Delete(_ context.Context, id uuid.UUID) error {
	if r.profile != nil && r.profile.ID == id {
		r.profile = nil
	}
	return nil
}

type accountPatientAccessRepository struct {
	repository.PatientAccessRepo
}

func (accountPatientAccessRepository) ListAccessiblePatientsByUser(context.Context, uuid.UUID, int, int) ([]repository.AccessiblePatient, int64, error) {
	return []repository.AccessiblePatient{}, 0, nil
}
