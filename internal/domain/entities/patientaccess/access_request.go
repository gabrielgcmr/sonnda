package patientaccess

import (
	"time"

	"github.com/google/uuid"
)

type RequestStatus string

const (
	RequestPending   RequestStatus = "pending"  // solicitado, aguardando ação
	RequestApproved  RequestStatus = "approved" // aprovado (opcional; pode ir direto para grant)
	RequestRejected  RequestStatus = "rejected"
	RequestCancelled RequestStatus = "cancelled"
	RequestExpired   RequestStatus = "expired"
)

type PatientAccessRequest struct {
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
