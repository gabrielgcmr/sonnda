package user

import (
	"errors"
)

var (
	ErrInvalidAuthProvider        = errors.New("auth provider is required")
	ErrInvalidAuthSubject         = errors.New("auth subject is required")
	ErrInvalidEmail               = errors.New("email is required or invalid")
	ErrInvalidRole                = errors.New("invalid user role")
	ErrInvalidBirthDate           = errors.New("invalid birth date")
	ErrInvalidFullName            = errors.New("full name is required")
	ErrInvalidPhone               = errors.New("phone is required")
	ErrInvalidCPF                 = errors.New("cpf is required")
	ErrUnsupportedBirthDateType   = errors.New("BirthDate.Scan: unsupported type")
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrCPFAlreadyExists           = errors.New("cpf already exists")
	ErrAuthorizationForbidden     = errors.New("authorization forbidden")
	ErrAuthIdentityAlreadyExists  = errors.New("auth identity already exists")
	ErrIdentityAlreadyLinkedError = errors.New("identity already linked to another user")
	ErrUserNotFound               = errors.New("user not found")
)
