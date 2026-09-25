// internal/features/patient/profile/service.go
package patientprofile

import (
	"context"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"

	"github.com/google/uuid"
)

type Service interface {
	Get(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) (*profiledomain.Patient, error)
	Update(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID, input UpdateInput) (*profiledomain.Patient, error)
	SoftDelete(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) error
	HardDelete(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) error
	ListMyPatients(ctx context.Context, currentUser *accountdomain.User, limit, offset int) ([]*profiledomain.Patient, error)
}

type AccessChecker interface {
	RequireAccess(ctx context.Context, accountID, patientID uuid.UUID) error
}
