package domain

import "time"

// Sistema de autorizações com histórico
type Authorization struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	AuthorizedID string     `gorm:"not null" json:"user_id"`
	PatientID    string     `gorm:"not null" json:"patient_id"`
	Status       AuthStatus `gorm:"type:varchar(20);not null" json:"status"`
	RequestedAt  time.Time  `json:"requested_at"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`

	// Histórico de alterações
	History []AuthorizationHistory `gorm:"foreignKey:AuthorizationID" json:"history"`
}

type AuthStatus string

const (
	AuthPending  AuthStatus = "PENDING"
	AuthApproved AuthStatus = "APPROVED"
	AuthRevoked  AuthStatus = "REVOKED"
	AuthExpired  AuthStatus = "EXPIRED"
)

type AuthorizationHistory struct {
	ID              string     `gorm:"primaryKey"`
	AuthorizationID string     `gorm:"not null"`
	OldStatus       AuthStatus `gorm:"type:varchar(20)"`
	NewStatus       AuthStatus `gorm:"type:varchar(20);not null"`
	ChangedBy       string     `gorm:"not null"` // UserID de quem mudou
	Reason          string     `json:"reason,omitempty"`
	ChangedAt       time.Time  `json:"changed_at"`
}
