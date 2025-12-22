package identity

import "errors"

var (
	ErrInvalidAuthProvider    = errors.New("auth provider is required")
	ErrInvalidAuthSubject     = errors.New("auth subject is required")
	ErrInvalidEmail           = errors.New("email is required or invalid")
	ErrInvalidRole            = errors.New("invalid user role")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrAuthorizationForbidden = errors.New("authorization forbidden")
)
