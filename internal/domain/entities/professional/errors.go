package professional

import "errors"

var (
	ErrRegistrationRequired      = errors.New("professional registration is required")
	ErrInvalidUserID             = errors.New("user id is required")
	ErrInvalidRegistrationNumber = errors.New("registration number is required")
	ErrInvalidRegistrationIssuer = errors.New("registration issuer is required")
	ErrInvalidStatus             = errors.New("status must be pending, verified, or rejected")
	ErrProfileNotFound           = errors.New("professional profile not found")
)
