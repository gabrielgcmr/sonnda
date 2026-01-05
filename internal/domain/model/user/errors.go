package user

import (
	"errors"
)

var (
	//invalid data errors
	ErrInvalidAuthProvider = errors.New("invalid auth provider")
	ErrInvalidAuthSubject  = errors.New("invalid auth subject")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidAccountType  = errors.New("invalid account type")
	ErrInvalidBirthDate    = errors.New("invalid birth date")
	ErrInvalidFullName     = errors.New("invalid full name")
	ErrInvalidPhone        = errors.New("invalid phone")
	ErrInvalidCPF          = errors.New("invalid cpf")
	//Erros não pertencem ao pacote /domain/user?

	ErrAuthIdentityAlreadyExists  = errors.New("auth identity already exists")
	ErrAuthorizationForbidden     = errors.New("authorization forbidden")
	ErrIdentityAlreadyLinkedError = errors.New("identity already linked to another user")
)
