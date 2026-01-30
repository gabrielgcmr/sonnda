// internal/domain/entity/professional/errors.go
package professional

import "errors"

var (
	ErrInvalidUserID             = errors.New("user id is required")
	ErrInvalidKind               = errors.New("invalid professional kind")
	ErrInvalidRegistrationNumber = errors.New("registration number is required")
	ErrInvalidRegistrationIssuer = errors.New("registration issuer is required")
	ErrInvalidStatus             = errors.New("status must be pending, verified, or rejected")
)
