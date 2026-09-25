// internal/features/patient/access/request_repository.go
package patientaccess

import (
	"context"

	accessdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/access/domain"

	"github.com/google/uuid"
)

// RequestRepository persists the future patient access request workflow.
type RequestRepository interface {
	Get(ctx context.Context, requestID uuid.UUID) (*accessdomain.AccessRequest, bool, error)
	Save(ctx context.Context, req accessdomain.AccessRequest) error
}
