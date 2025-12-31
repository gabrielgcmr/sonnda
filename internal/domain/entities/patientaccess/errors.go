package patientaccess

import "errors"

var (
	ErrInvalidPatientID = errors.New("patient id is required")
	ErrInvalidUserID    = errors.New("app user id is required")
	ErrInvalidRole      = errors.New("role is required")
)
