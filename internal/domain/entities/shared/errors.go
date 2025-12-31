package shared

import "errors"

var (
	ErrInvalidBirthDate         = errors.New("invalid birth date")
	ErrInvalidCPF               = errors.New("invalid cpf")
	ErrInvalidFullName          = errors.New("full name is required")
	ErrInvalidPhone             = errors.New("phone is required")
	ErrUnsupportedBirthDateType = errors.New("BirthDate.Scan: unsupported type")
)
