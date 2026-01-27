// internal/adapters/outbound/integrations/auth/firebase_auth_service.go
package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"github.com/gabrielgcmr/sonnda/internal/domain/model/identity"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports"
)

type FirebaseAuthService struct {
	client    *firebaseauth.Client
	projectID string
}

var _ ports.IdentityService = (*FirebaseAuthService)(nil)

func NewFirebaseAuthService(ctx context.Context) (*FirebaseAuthService, error) {
	projectID := strings.TrimSpace(os.Getenv("GCP_PROJECT_ID"))

	var opts []option.ClientOption
	if credentialsJSON := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_JSON")); credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	} else if credentialsFile := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_FILE")); credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}

	// ADC:
	// - Local: GOOGLE_APPLICATION_CREDENTIALS aponta para o arquivo JSON
	// - Cloud Run: Service Account do serviço fornece credenciais via metadata server
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase init: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("app.auth: %w", err)
	}

	return &FirebaseAuthService{client: authClient, projectID: projectID}, nil
}

func (s *FirebaseAuthService) ProviderName() string {
	return "firebase"
}

func (s *FirebaseAuthService) VerifyToken(ctx context.Context, tokenStr string) (*identity.Identity, error) {
	if s.client == nil {
		return nil, errors.New("firebase client not configured")
	}

	token, err := s.client.VerifyIDToken(ctx, tokenStr)
	if err != nil {
		return nil, errors.New("invalid bearer token")
	}

	email, _ := token.Claims["email"].(string)

	return &identity.Identity{
		Provider: "firebase",
		Subject:  token.UID,
		Email:    email,
	}, nil
}

func (s *FirebaseAuthService) VerifySessionCookie(ctx context.Context, sessionCookie string) (*identity.Identity, error) {
	if s.client == nil {
		return nil, errors.New("firebase client not configured")
	}

	token, err := s.client.VerifySessionCookieAndCheckRevoked(ctx, sessionCookie)
	if err != nil {
		return nil, errors.New("invalid session cookie")
	}

	email, _ := token.Claims["email"].(string)

	return &identity.Identity{
		Provider: "firebase",
		Subject:  token.UID,
		Email:    email,
	}, nil
}

func (s *FirebaseAuthService) CreateSessionCookie(ctx context.Context, idToken string, expiresIn time.Duration) (string, error) {
	if s.client == nil {
		return "", errors.New("firebase client not configured")
	}

	sessionCookie, err := s.client.SessionCookie(ctx, idToken, expiresIn)
	if err != nil {
		return "", fmt.Errorf("firebase session cookie: %w", err)
	}

	return sessionCookie, nil
}

func (s *FirebaseAuthService) RevokeSessions(ctx context.Context, subject string) error {
	if s.client == nil {
		return errors.New("firebase client not configured")
	}

	if err := s.client.RevokeRefreshTokens(ctx, subject); err != nil {
		return fmt.Errorf("firebase revoke refresh tokens: %w", err)
	}

	return nil
}

func (s *FirebaseAuthService) DisableUser(ctx context.Context, subject string) error {
	if s.client == nil {
		return errors.New("firebase client not configured")
	}

	_, err := s.client.UpdateUser(
		ctx,
		subject,
		(&firebaseauth.UserToUpdate{}).Disabled(true),
	)
	if err != nil {
		return fmt.Errorf("firebase disable user: %w", err)
	}

	return nil
}
