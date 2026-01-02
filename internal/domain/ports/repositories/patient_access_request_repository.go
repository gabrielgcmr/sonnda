package repositories

import (
	"context"
	"sonnda-api/internal/domain/entities/patientaccess"

	"github.com/google/uuid"
)

// RequestRepository é a porta de domínio para persistir PatientAccessRequest (workflow).
type RequestRepository interface {
	Get(ctx context.Context, requestID uuid.UUID) (*patientaccess.AccessRequest, bool, error)
	Save(ctx context.Context, req patientaccess.AccessRequest) error
}
