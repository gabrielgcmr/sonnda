// internal/shared/security/identity.go
package security

// Identity representa o usuário autenticado (OIDC-compatible).
type Identity struct {
	Issuer  string
	Subject string

	Email         *string
	EmailVerified *bool

	Name       *string
	PictureURL *string

	Scopes []string
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
