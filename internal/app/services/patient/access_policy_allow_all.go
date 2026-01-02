// internal/app/services/patient/access_policy_allow_all.go
package patientsvc

import (
	"context"

	"sonnda-api/internal/domain/model/user"

	"github.com/google/uuid"
)

type AllowAllPolicy struct{}

func (p AllowAllPolicy) CanCreate(ctx context.Context, currentUser *user.User, input CreateInput) error {
	return nil
}
func (p AllowAllPolicy) CanRead(ctx context.Context, currentUser *user.User, patientID uuid.UUID) error {
	return nil
}
func (p AllowAllPolicy) CanUpdate(ctx context.Context, currentUser *user.User, patientID uuid.UUID) error {
	return nil
}
func (p AllowAllPolicy) CanDelete(ctx context.Context, currentUser *user.User, patientID uuid.UUID) error {
	return nil
}
func (p AllowAllPolicy) CanList(ctx context.Context, currentUser *user.User) error {
	return nil
}
