package identity

import "time"

type Authorization struct {
	ID           string
	AuthorizedID string
	PatientID    string
	Status       AuthStatus
	RequestedAt  time.Time
	ApprovedAt   *time.Time
	RevokedAt    *time.Time
	History      []AuthorizationHistory
}

type AuthStatus string

const (
	AuthPending  AuthStatus = "PENDING"
	AuthApproved AuthStatus = "APPROVED"
	AuthRevoked  AuthStatus = "REVOKED"
	AuthExpired  AuthStatus = "EXPIRED"
)

type AuthorizationHistory struct {
	ID              string
	AuthorizationID string
	OldStatus       AuthStatus
	NewStatus       AuthStatus
	ChangedBy       string
	Reason          string
	ChangedAt       time.Time
}
