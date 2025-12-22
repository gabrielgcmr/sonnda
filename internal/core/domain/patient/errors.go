package patient

import "errors"

var (
	ErrPatientNotFound  = errors.New("patient not found")
	ErrCPFAlreadyExists = errors.New("cpf already exists")
	ErrInvalidBirthDate = errors.New("invalid birth date")
)
