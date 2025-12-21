// internal/adapters/outbound/auth/firebase_auth_service.go
package auth

import (
	"context"
	"errors"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"sonnda-api/internal/core/ports/services"
)

type FirebaseConfig struct {
	ProjectID       string
	CredentialsFile string
}

type FirebaseAuthService struct {
	client *firebaseauth.Client
}

var _ services.AuthService = (*FirebaseAuthService)(nil)

func NewFirebaseAuthService(ctx context.Context, cfg FirebaseConfig) (*FirebaseAuthService, error) {
	if cfg.CredentialsFile == "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return nil, errors.New("firebase credentials not configured")
	}

	var appOpts []option.ClientOption
	if cfg.CredentialsFile != "" {
		appOpts = append(appOpts, option.WithCredentialsFile(cfg.CredentialsFile))
	}

	var firebaseCfg *firebase.Config
	if cfg.ProjectID != "" {
		firebaseCfg = &firebase.Config{ProjectID: cfg.ProjectID}
	}

	app, err := firebase.NewApp(ctx, firebaseCfg, appOpts...)
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase auth: %w", err)
	}

	return &FirebaseAuthService{client: client}, nil
}

func (s *FirebaseAuthService) VerifyToken(ctx context.Context, tokenStr string) (*services.Identity, error) {
	if tokenStr == "" {
		return nil, errors.New("missing token")
	}

	token, err := s.client.VerifyIDToken(ctx, tokenStr)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	email := ""
	if rawEmail, ok := token.Claims["email"]; ok {
		if val, ok := rawEmail.(string); ok {
			email = val
		}
	}
	if email == "" {
		return nil, errors.New("token missing email")
	}

	return &services.Identity{
		Provider: "firebase",
		Subject:  token.UID,
		Email:    email,
	}, nil
}
