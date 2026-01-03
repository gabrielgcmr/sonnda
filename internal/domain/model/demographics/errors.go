package demographics

import "errors"

var (
	ErrInvalidBirthDate         = errors.New("invalid birth date")
	ErrInvalidCPF               = errors.New("invalid cpf")
	ErrInvalidFullName          = errors.New("full name is required")
	ErrInvalidGender            = errors.New("invalid gender")
	ErrInvalidPhone             = errors.New("phone is required")
	ErrInvalidRace              = errors.New("invalid race")
	ErrUnsupportedBirthDateType = errors.New("BirthDate.Scan: unsupported type")
)
