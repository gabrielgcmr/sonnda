package patientaccess

import (
	"time"

	"github.com/google/uuid"
)

//Não foi criado um status para o vínculo, pois ele se ele existe já é ativo ou não ser que tenha sido explicitamente revogado.

type PatientAccess struct {
	PatientID uuid.UUID
	UserID    uuid.UUID

	Type RelationshipType

	CreatedAt time.Time
	RevokedAt *time.Time //Se nulo, está ativo
	GrantedBy *uuid.UUID
}

func NewPatientAccess(
	patientID, userID uuid.UUID,
	relType RelationshipType,
	grantedBy *uuid.UUID,
	now time.Time,
) (*PatientAccess, error) {
	if patientID == uuid.Nil {
		return nil, ErrInvalidPatientID
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if now.IsZero() {
		return nil, ErrInvalidTimestamp
	}

	if !relType.IsValid() {
		return nil, ErrInvalidRelationshipType
	}

	pa := &PatientAccess{
		PatientID: patientID,
		UserID:    userID,
		Type:      relType,
		CreatedAt: now,
		RevokedAt: nil,
		GrantedBy: grantedBy,
	}

	return pa, nil
}

func (pa *PatientAccess) Validate() error {
	if pa == nil {
		return ErrNilPatientAccess
	}

	if pa.PatientID == uuid.Nil {
		return ErrInvalidPatientID
	}
	if pa.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	if !pa.Type.IsValid() {
		return ErrInvalidRelationshipType
	}
	if pa.CreatedAt.IsZero() {
		return ErrInvalidTimestamp
	}
	if pa.RevokedAt != nil && pa.RevokedAt.Before(pa.CreatedAt) {
		return ErrRevokedBeforeCreated
	}
	return nil
}

func (pa *PatientAccess) IsActive() bool {
	return pa.RevokedAt == nil
}

func (pa *PatientAccess) IsRevoked() bool {
	return pa.RevokedAt != nil
}

func (pa *PatientAccess) Revoke(revokedAt time.Time) error {
	if pa == nil {
		return ErrNilPatientAccess
	}
	if !pa.IsActive() {
		return ErrAccessAlreadyRevoked
	}
	if revokedAt.IsZero() {
		return ErrInvalidTimestamp
	}
	if revokedAt.Before(pa.CreatedAt) {
		return ErrRevokedBeforeCreated
	}
	pa.RevokedAt = &revokedAt
	return nil
}
