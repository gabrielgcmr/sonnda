// internal/features/patient/profile/service.go
package patientprofile

import (
	"context"

	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	profiledomain "github.com/gabrielgcmr/sonnda/internal/features/patient/profile/domain"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, currentUser *accountdomain.User, input CreateInput) (*profiledomain.Patient, error)
	Get(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) (*profiledomain.Patient, error)
	Update(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID, input UpdateInput) (*profiledomain.Patient, error)
	SoftDelete(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) error
	HardDelete(ctx context.Context, currentUser *accountdomain.User, id uuid.UUID) error
	ListMyPatients(ctx context.Context, currentUser *accountdomain.User, limit, offset int) ([]*profiledomain.Patient, error)
}

type Authorizer interface {
	RequirePatientAccess(ctx context.Context, actor *accountdomain.User, patientID uuid.UUID) error
}
