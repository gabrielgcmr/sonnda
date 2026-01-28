package user

import (
	"errors"
)

var (
	//invalid data errors
	ErrInvalidAuthIssuer  = errors.New("invalid auth provider")
	ErrInvalidAuthSubject = errors.New("invalid auth subject")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidAccountType = errors.New("invalid account type")
	ErrInvalidBirthDate   = errors.New("invalid birth date")
	ErrInvalidFullName    = errors.New("invalid full name")
	ErrInvalidPhone       = errors.New("invalid phone")
	ErrInvalidCPF         = errors.New("invalid cpf")
)
