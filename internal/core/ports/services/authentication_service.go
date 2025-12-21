// internal/core/ports/services/auth_service.go
package services

import (
	"context"
)

type Identity struct {
	Provider string // ex: "firebase"
	Subject  string // firebase uid
	Email    string
}

type AuthService interface {
	VerifyToken(ctx context.Context, token string) (*Identity, error)
}
