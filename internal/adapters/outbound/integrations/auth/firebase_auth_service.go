// internal/infrastructure/auth/firebase_auth_service.go
package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"sonnda-api/internal/domain/model/identity"
	"sonnda-api/internal/domain/ports/integration"
)

type FirebaseAuthService struct {
	client *firebaseauth.Client
}

func NewFirebaseAuthService(ctx context.Context) (*FirebaseAuthService, error) {
	var opts []option.ClientOption

	credentialsJSON := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_JSON"))
	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	} else {
		credentialsFile := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_FILE"))
		if credentialsFile != "" {
			opts = append(opts, option.WithCredentialsFile(credentialsFile))
		}
	}

	projectID := strings.TrimSpace(os.Getenv("FIREBASE_PROJECT_ID"))
	if projectID == "" {
		projectID = strings.TrimSpace(os.Getenv("GCP_PROJECT_ID"))
	}

	var fbConfig *firebase.Config
	if projectID != "" {
		fbConfig = &firebase.Config{ProjectID: projectID}
	}

	app, err := firebase.NewApp(ctx, fbConfig, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase init: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase auth: %w", err)
	}

	return &FirebaseAuthService{client: client}, nil
}

var _ integration.IdentityService = (*FirebaseAuthService)(nil)

func (s *FirebaseAuthService) ProviderName() string {
	return "firebase"
}

func (s *FirebaseAuthService) VerifyToken(ctx context.Context, tokenStr string) (*identity.Identity, error) {
	if s.client == nil {
		return nil, errors.New("firebase client not configured")
	}

	token, err := s.client.VerifyIDToken(ctx, tokenStr)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	email, _ := token.Claims["email"].(string)

	return &identity.Identity{
		Provider: "firebase",
		Subject:  token.UID,
		Email:    email,
	}, nil
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
