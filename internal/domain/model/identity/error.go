package identity

import (
	"errors"
)

var (
	//invalid data errors
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrCPFAlreadyExists           = errors.New("cpf already exists")
	ErrAuthorizationForbidden     = errors.New("authorization forbidden")
	ErrAuthIdentityAlreadyExists  = errors.New("auth identity already exists")
	ErrIdentityAlreadyLinkedError = errors.New("identity already linked to another user")
	ErrUserNotFound               = errors.New("user not found")
)
