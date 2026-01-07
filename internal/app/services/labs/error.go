package labsvc

import "errors"

var (
	ErrAuthorizationForbidden = errors.New("authorization forbidden")
	ErrPatientNotFound        = errors.New("patient not found")
	ErrCPFAlreadyExists       = errors.New("cpf already exists")
)
