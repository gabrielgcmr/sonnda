package patientaccess

import "errors"

var (
	// Common
	ErrInvalidTimestamp = errors.New("invalid timestamp provided")
	ErrInvalidPatientID = errors.New("patient id is required")

	// PatientAccess specific
	ErrRevokedBeforeCreated = errors.New("revokedAt cannot be before createdAt")
	ErrInvalidUserID        = errors.New("user id is required")

	// AccessRequest specific
	ErrNilPatientAccess         = errors.New("nil patient access")
	ErrInvalidDeciderUserID     = errors.New("decider user id is required")
	ErrInvalidRequesterUserID   = errors.New("requester user id is required")
	ErrInvalidTargetUserID      = errors.New("target user id is required")
	ErrExpiresAtBeforeCreatedAt = errors.New("expiresAt must be after createdAt")
	ErrNilAccessRequest         = errors.New("nil access request")
	ErrRequestNotPending        = errors.New("patient access request is not pending")
	ErrRequestExpired           = errors.New("patient access request has expired")
	ErrRequestAlreadyDecided    = errors.New("patient access request has already been decided")
)
