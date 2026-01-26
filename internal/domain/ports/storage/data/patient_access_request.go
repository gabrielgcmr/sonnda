package data

import (
	"context"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/patientaccess"

	"github.com/google/uuid"
)

// RequestRepo é a porta de domínio para persistir PatientAccessRequest (workflow).
type RequestRepo interface {
	Get(ctx context.Context, requestID uuid.UUID) (*patientaccess.AccessRequest, bool, error)
	Save(ctx context.Context, req patientaccess.AccessRequest) error
}
