package patient

import "errors"

var (
	ErrInvalidFullName  = errors.New("invalid full name")
	ErrInvalidBirthDate = errors.New("invalid birth date")
	ErrInvalidGender    = errors.New("invalid gender")
	ErrInvalidRace      = errors.New("invalid race")
)
