// internal/domain/entity/medicalrecord/errors.go
package medicalrecord

import "errors"

var (
	ErrMedicalRecordNotFound = errors.New("medical record not found")
	ErrInvalidRecordID       = errors.New("medical record id is required")
	ErrInvalidPatientID      = errors.New("patient id is required")
)
