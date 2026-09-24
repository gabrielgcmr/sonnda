// internal/features/account/middleware_test.go
package account

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/api/helpers"
	"github.com/gabrielgcmr/sonnda/internal/api/presenter"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/identity"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/gin-gonic/gin"
)

type registrationRepo struct {
	repository.User
	profile *user.User
	err     error
}

func (r registrationRepo) FindByAuthIdentity(context.Context, string, string) (*user.User, error) {
	return r.profile, r.err
}

func TestRegistrationAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name          string
		authenticated bool
		profile       *user.User
		err           error
		status        int
		code          apperr.ErrorKind
	}{
		{"no session", false, nil, nil, 401, apperr.AUTH_REQUIRED},
		{"missing profile", true, nil, nil, 403, apperr.PROFILE_NOT_FOUND},
		{"database failure", true, nil, errors.New("private database details"), 500, apperr.INTERNAL_ERROR},
		{"registered", true, &user.User{FullName: "Ana"}, nil, 200, ""},
	} {
		for _, path := range []string{"/v1/me", "/v1/patients"} {
			t.Run(tc.name+path, func(t *testing.T) {
				router := gin.New()
				router.Use(func(c *gin.Context) {
					if tc.authenticated {
						helpers.SetIdentity(c, &identity.Identity{Issuer: "test", Subject: "user-1"})
					}
				})
				registration := NewMiddleware(registrationRepo{profile: tc.profile, err: tc.err})
				reached := false
				router.GET(path, registration.RequireRegisteredUser(), func(c *gin.Context) {
					reached = true
					if helpers.MustGetCurrentUser(c) != tc.profile {
						t.Fatal("incorrect profile")
					}
					c.Status(http.StatusOK)
				})
				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
				if response.Code != tc.status {
					t.Fatalf("status = %d, want %d", response.Code, tc.status)
				}
				if reached != (tc.status == 200) {
					t.Fatal("protected handler access mismatch")
				}
				if tc.code != "" {
					var problem presenter.Problem
					if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
						t.Fatal(err)
					}
					if problem.Code != tc.code {
						t.Fatalf("code = %s, want %s", problem.Code, tc.code)
					}
				}
			})
		}
	}
}

func TestMissingProfileIsExpectedClientError(t *testing.T) {
	err := apperr.ProfileNotFound()
	if !apperr.IsForbidden(err) || apperr.LogLevelOf(err) != slog.LevelInfo {
		t.Fatal("missing profile must remain an expected authorization error")
	}
}
