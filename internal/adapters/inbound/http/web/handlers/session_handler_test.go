// internal/adapters/inbound/http/web/handlers/session_handler_test.go
package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gabrielgcmr/sonnda/internal/domain/model/identity"
)

type fakeIdentityService struct {
	verifySessionCookie func(ctx context.Context, sessionCookie string) (*identity.Identity, error)
	revokeSessions      func(ctx context.Context, subject string) error
}

func (f *fakeIdentityService) ProviderName() string { return "fake" }
func (f *fakeIdentityService) VerifyToken(ctx context.Context, tokenStr string) (*identity.Identity, error) {
	panic("not used")
}
func (f *fakeIdentityService) VerifySessionCookie(ctx context.Context, sessionCookie string) (*identity.Identity, error) {
	if f.verifySessionCookie == nil {
		panic("VerifySessionCookie not configured")
	}
	return f.verifySessionCookie(ctx, sessionCookie)
}
func (f *fakeIdentityService) CreateSessionCookie(ctx context.Context, idToken string, expiresIn time.Duration) (string, error) {
	panic("not used")
}
func (f *fakeIdentityService) RevokeSessions(ctx context.Context, subject string) error {
	if f.revokeSessions == nil {
		panic("RevokeSessions not configured")
	}
	return f.revokeSessions(ctx, subject)
}
func (f *fakeIdentityService) DisableUser(ctx context.Context, subject string) error { panic("not used") }

func TestSessionHandler_Logout_NavigateRedirectsEvenWithoutCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSessionHandler(&fakeIdentityService{})
	r := gin.New()
	r.POST("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Sec-Fetch-Mode", "navigate")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/login" {
		t.Fatalf("expected Location /login, got %q", loc)
	}
}

func TestSessionHandler_Logout_FetchReturnsNoContentWhenNoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSessionHandler(&fakeIdentityService{})
	r := gin.New()
	r.POST("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Sec-Fetch-Mode", "cors")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "" {
		t.Fatalf("expected no Location header, got %q", loc)
	}
}

func TestSessionHandler_Logout_RevokeSessionsAndRedirectOnNavigate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var revokedSubject string
	fake := &fakeIdentityService{
		verifySessionCookie: func(ctx context.Context, sessionCookie string) (*identity.Identity, error) {
			if sessionCookie != "cookie123" {
				t.Fatalf("unexpected cookie %q", sessionCookie)
			}
			return &identity.Identity{Provider: "fake", Subject: "user-1"}, nil
		},
		revokeSessions: func(ctx context.Context, subject string) error {
			revokedSubject = subject
			return nil
		},
	}

	h := NewSessionHandler(fake)
	r := gin.New()
	r.POST("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.AddCookie(&http.Cookie{Name: firebaseSessionCookieName, Value: "cookie123"})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if revokedSubject != "user-1" {
		t.Fatalf("expected revoked subject %q, got %q", "user-1", revokedSubject)
	}
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/login" {
		t.Fatalf("expected Location /login, got %q", loc)
	}
}

