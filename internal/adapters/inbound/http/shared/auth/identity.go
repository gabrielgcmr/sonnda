// internal/adapters/inbound/http/shared/auth/identity.go
package auth

import "time"

// Identity representa o usuário autenticado (OIDC-compatible).
type Identity struct {
	Issuer  string
	Subject string

	Provider *string

	Email         *string
	EmailVerified *bool

	Name       *string
	PictureURL *string

	AuthTime *time.Time
	Scopes   []string
	Claims   map[string]any
}

// PrincipalID retorna o identificador canônico do principal.
// Derivado de Issuer+Subject para evitar inconsistência.
func (i Identity) PrincipalID() string {
	return i.Issuer + "|" + i.Subject
}

// NewIdentity cria uma Identity com PrincipalID canônico.
func NewIdentity(issuer, subject string) Identity {
	return Identity{
		Issuer:  issuer,
		Subject: subject,
	}
}
