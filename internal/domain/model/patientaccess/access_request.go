package patientaccess

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AccessRequest struct {
	ID        uuid.UUID
	PatientID uuid.UUID

	RequesterUserID       uuid.UUID        // quem pediu (ex.: cardiologista)
	RequesterRelationType RelationshipType // profissional/family/caregiver (ou algo mais específico)

	TargetUserID       uuid.UUID        // quem terá acesso (às vezes é o mesmo do requester)
	TargetRelationType RelationshipType // professional/family/caregiver (ou algo mais específico)

	Status    RequestStatus
	Reason    *string // opcional: justificativa
	CreatedAt time.Time
	ExpiresAt *time.Time

	DecidedBy *uuid.UUID // quem aprovou/rejeitou (paciente/self, cuidador autorizado, etc.)
	DecidedAt *time.Time
}

func NewAccessRequest(
	patientID uuid.UUID,
	requesterUserID uuid.UUID,
	requesterRelationType RelationshipType,
	targetUserID uuid.UUID,
	targetRelationType RelationshipType,
	expiresAt *time.Time,
	reason *string,
	now time.Time,
) (*AccessRequest, error) {
	if patientID == uuid.Nil {
		return nil, ErrInvalidPatientID
	}
	if requesterUserID == uuid.Nil {
		return nil, ErrInvalidRequesterUserID
	}
	if targetUserID == uuid.Nil {
		return nil, ErrInvalidTargetUserID
	}
	if now.IsZero() {
		return nil, ErrInvalidTimestamp
	}
	if !requesterRelationType.IsValid() || !targetRelationType.IsValid() {
		return nil, ErrInvalidRelationshipType
	}
	if expiresAt != nil && expiresAt.Before(now) {
		// Se você quiser permitir “expira imediatamente”, troque para Before/Equal conforme regra.
		return nil, ErrExpiresAtBeforeCreatedAt
	}

	ar := &AccessRequest{
		ID:                    uuid.New(),
		PatientID:             patientID,
		RequesterUserID:       requesterUserID,
		RequesterRelationType: requesterRelationType,
		TargetUserID:          targetUserID,
		TargetRelationType:    targetRelationType,
		Status:                RequestPending,
		Reason:                reason,
		CreatedAt:             now,
		ExpiresAt:             expiresAt,
		DecidedBy:             nil,
		DecidedAt:             nil,
	}

	return ar, nil
}

func (ar *AccessRequest) Validate() error {
	if ar == nil {
		return ErrNilAccessRequest
	}
	if ar.ID == uuid.Nil {
		return errors.New("access request id is required")
	}
	if ar.PatientID == uuid.Nil {
		return ErrInvalidPatientID
	}
	if ar.RequesterUserID == uuid.Nil {
		return ErrInvalidRequesterUserID
	}
	if ar.TargetUserID == uuid.Nil {
		return ErrInvalidTargetUserID
	}
	if ar.CreatedAt.IsZero() {
		return ErrInvalidTimestamp
	}
	if !ar.Status.IsValid() {
		return errors.New("invalid access request status")
	}
	if !ar.RequesterRelationType.IsValid() || !ar.TargetRelationType.IsValid() {
		return ErrInvalidRelationshipType
	}
	if ar.ExpiresAt != nil && ar.ExpiresAt.Before(ar.CreatedAt) {
		return ErrExpiresAtBeforeCreatedAt
	}

	// Consistência entre Status e campos de decisão
	if ar.Status == RequestPending || ar.Status == RequestExpired || ar.Status == RequestCancelled {
		// nesses estados, decisão humana não é obrigatória
		return nil
	}
	// approved/rejected: deve ter decisão registrada
	if ar.DecidedBy == nil || ar.DecidedAt == nil {
		return errors.New("missing decision fields")
	}
	return nil
}

func (ar *AccessRequest) IsPending() bool {
	return ar != nil && ar.Status == RequestPending
}

func (ar *AccessRequest) IsExpired(now time.Time) bool {
	if ar == nil || ar.ExpiresAt == nil {
		return false
	}
	return now.After(*ar.ExpiresAt)
}

// ExpireIfNeeded é idempotente: se já não estiver pending, não faz nada.
// Se estiver pending e já passou do ExpiresAt, vira expired.
func (ar *AccessRequest) ExpireIfNeeded(now time.Time) bool {
	if ar == nil || now.IsZero() {
		return false
	}
	if ar.Status != RequestPending {
		return false
	}
	if ar.ExpiresAt == nil {
		return false
	}
	if !now.After(*ar.ExpiresAt) {
		return false
	}
	ar.Status = RequestExpired
	decidedAt := now
	ar.DecidedAt = &decidedAt
	return true
}

func (ar *AccessRequest) Approve(deciderUserID uuid.UUID, now time.Time) error {
	if ar == nil {
		return ErrNilAccessRequest
	}
	if deciderUserID == uuid.Nil {
		return ErrInvalidDeciderUserID
	}
	if now.IsZero() {
		return ErrInvalidTimestamp
	}

	ar.ExpireIfNeeded(now)
	if ar.Status == RequestExpired {
		return ErrRequestExpired
	}
	if ar.Status != RequestPending {
		return ErrRequestNotPending
	}
	if ar.DecidedAt != nil || ar.DecidedBy != nil {
		return ErrRequestAlreadyDecided
	}

	ar.Status = RequestApproved
	ar.DecidedBy = &deciderUserID
	ar.DecidedAt = &now
	return nil
}

func (ar *AccessRequest) Reject(deciderUserID uuid.UUID, now time.Time, reason *string) error {
	if ar == nil {
		return ErrNilAccessRequest
	}
	if deciderUserID == uuid.Nil {
		return ErrInvalidDeciderUserID
	}
	if now.IsZero() {
		return ErrInvalidTimestamp
	}

	ar.ExpireIfNeeded(now)
	if ar.Status == RequestExpired {
		return ErrRequestExpired
	}
	if ar.Status != RequestPending {
		return ErrRequestNotPending
	}
	if ar.DecidedAt != nil || ar.DecidedBy != nil {
		return ErrRequestAlreadyDecided
	}

	ar.Status = RequestRejected
	ar.DecidedBy = &deciderUserID
	ar.DecidedAt = &now
	ar.Reason = reason
	return nil
}

func (ar *AccessRequest) Cancel(now time.Time) error {
	if ar == nil {
		return ErrNilAccessRequest
	}
	if now.IsZero() {
		return ErrInvalidTimestamp
	}

	ar.ExpireIfNeeded(now)
	if ar.Status == RequestExpired {
		return ErrRequestExpired
	}
	if ar.Status != RequestPending {
		return ErrRequestNotPending
	}

	ar.Status = RequestCancelled
	return nil
}
