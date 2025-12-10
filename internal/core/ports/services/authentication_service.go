// internal/core/ports/services/auth_service.go
package services

import (
	"context"
)

type Identity struct {
	Provider string // ex: "supabase"
	Subject  string // auth.users.id
	Email    string
}

type AuthService interface {
	VerifyToken(ctx context.Context, token string) (*Identity, error)
}
