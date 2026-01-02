package patientaccess

import "errors"

var (
	// Validation / input
	ErrInvalidPatientID      = errors.New("patient id is required")
	ErrInvalidUserID         = errors.New("user id is required")
	ErrInvalidRequesterUserID = errors.New("requester user id is required")
	ErrInvalidTargetUserID    = errors.New("target user id is required")
	ErrInvalidDeciderUserID   = errors.New("decider user id is required")
	ErrInvalidRequestID       = errors.New("access request id is required")
	ErrInvalidRequestStatus   = errors.New("invalid access request status")
	ErrExpiresAtBeforeCreatedAt = errors.New("expiresAt must be after createdAt")
	ErrMissingDecisionFields    = errors.New("missing decision fields")
	ErrNilAccessRequest         = errors.New("nil access request")

	ErrAccessAlreadyExists                   = errors.New("patient access already exists")
	ErrAccessAlreadyRevoked                  = errors.New("patient access already revoked")
	ErrCannotRevokeLastActiveAccess          = errors.New("cannot revoke the last active patient access")
	ErrRequestNotPending                     = errors.New("patient access request is not pending")
	ErrRequestExpired                        = errors.New("patient access request has expired")
	ErrRequestAlreadyDecided                 = errors.New("patient access request has already been decided")
	ErrInvalidTransition                     = errors.New("invalid patient access request status transition")
	ErrDuplicateActiveAccess                 = errors.New("duplicate active patient access exists")
	ErrRequestNotApproved                    = errors.New("patient access request is not approved")
	ErrInvalidRelationshipType               = errors.New("invalid relationship type")
	ErrInvalidID                             = errors.New("invalid ID provided")
	ErrAccessNotFound                        = errors.New("patient access not found")
	ErrRequestNotFound                       = errors.New("patient access request not found")
	ErrInvalidTimestamp                      = errors.New("invalid timestamp provided")
	ErrRevokedBeforeCreated                  = errors.New("cannot revoke before creation")
	ErrPatientMustHaveAtLeastOneActiveAccess = errors.New("patient must have at least one active access")
	ErrNilPatientAccess                      = errors.New("nil patient access")
)
