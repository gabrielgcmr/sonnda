package patient

import "errors"

var (
	ErrPatientNotFound  = errors.New("patient not found")
	ErrCPFAlreadyExists = errors.New("cpf already exists")
	ErrInvalidFullName  = errors.New("invalid full name")
	ErrInvalidBirthDate = errors.New("invalid birth date")
	ErrInvalidGender    = errors.New("invalid gender")
	ErrInvalidRace      = errors.New("invalid race")
)
