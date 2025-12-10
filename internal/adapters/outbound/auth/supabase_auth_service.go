// internal/adapters/outbound/auth/supabase_auth_service.go
package auth

import (
	"context"
	"errors"
	"os"

	"sonnda-api/internal/core/ports/services"

	"github.com/golang-jwt/jwt/v5"
)

type SupabaseAuthService struct {
	jwtSecret []byte
}

func NewSupabaseAuthService() *SupabaseAuthService {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	return &SupabaseAuthService{
		jwtSecret: []byte(secret),
	}
}

type supabaseClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

var _ services.AuthService = (*SupabaseAuthService)(nil)

func (s *SupabaseAuthService) VerifyToken(ctx context.Context, tokenStr string) (*services.Identity, error) {
	if len(s.jwtSecret) == 0 {
		return nil, errors.New("supabase jwt secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &supabaseClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*supabaseClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return &services.Identity{
		Provider: "supabase",
		Subject:  claims.Subject, // sub = auth.users.id
		Email:    claims.Email,
	}, nil
}
