// internal/domain/entity/professional/professional_test.go
package professional

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewProfessionalProfile_Success_DefaultsAndTrims(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	state := "SP"

	p, err := NewProfessional(NewProfessionalParams{
		UserID:             userID,
		Kind:               KindDoctor,
		RegistrationNumber: "  CRM-123  ",
		RegistrationIssuer: "  CRM  ",
		RegistrationState:  &state,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if p == nil {
		t.Fatalf("expected profile, got nil")
	}
	if p.UserID != userID {
		t.Fatalf("expected user id, got %q", p.UserID)
	}
	if p.RegistrationNumber != "CRM-123" {
		t.Fatalf("expected registration number trimmed")
	}
	if p.RegistrationIssuer != "CRM" {
		t.Fatalf("expected registration issuer trimmed")
	}
	if p.RegistrationState == nil || *p.RegistrationState != "SP" {
		t.Fatalf("expected registration state trimmed")
	}
	if p.Status != StatusPending {
		t.Fatalf("expected status pending")
	}
	if p.VerifiedAt != nil {
		t.Fatalf("expected VerifiedAt to be nil")
	}
	if p.CreatedAt.Location() != time.UTC {
		t.Fatalf("expected CreatedAt in UTC")
	}
	if p.UpdatedAt.Location() != time.UTC {
		t.Fatalf("expected UpdatedAt in UTC")
	}
}

func TestNewProfessionalProfile_InvalidInputs(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cases := []struct {
		name string
		err  error
		fn   func() error
	}{
		{
			name: "missing user id",
			err:  ErrInvalidUserID,
			fn: func() error {
				_, err := NewProfessional(NewProfessionalParams{
					Kind:               KindDoctor,
					RegistrationNumber: "CRM-123",
					RegistrationIssuer: "CRM",
				})
				return err
			},
		},
		{
			name: "invalid kind",
			err:  ErrInvalidKind,
			fn: func() error {
				_, err := NewProfessional(NewProfessionalParams{
					UserID:             userID,
					Kind:               Kind(""),
					RegistrationNumber: "CRM-123",
					RegistrationIssuer: "CRM",
				})
				return err
			},
		},
		{
			name: "missing registration number",
			err:  ErrInvalidRegistrationNumber,
			fn: func() error {
				_, err := NewProfessional(NewProfessionalParams{
					UserID:             userID,
					Kind:               KindDoctor,
					RegistrationNumber: "   ",
					RegistrationIssuer: "CRM",
				})
				return err
			},
		},
		{
			name: "missing registration issuer",
			err:  ErrInvalidRegistrationIssuer,
			fn: func() error {
				_, err := NewProfessional(NewProfessionalParams{
					UserID:             userID,
					Kind:               KindDoctor,
					RegistrationNumber: "CRM-123",
					RegistrationIssuer: "   ",
				})
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected %v, got %v", tc.err, err)
			}
		})
	}
}

func TestProfile_SetStatus_VerifiedSetsVerifiedAt(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	p, err := NewProfessional(NewProfessionalParams{
		UserID:             userID,
		Kind:               KindDoctor,
		RegistrationNumber: "CRM-123",
		RegistrationIssuer: "CRM",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	before := p.UpdatedAt
	time.Sleep(2 * time.Millisecond)

	p.Verify()
	if p.Status != StatusVerified {
		t.Fatalf("expected status verified")
	}
	if p.VerifiedAt == nil {
		t.Fatalf("expected VerifiedAt to be set")
	}
	if !p.UpdatedAt.After(before) {
		t.Fatalf("expected UpdatedAt to move forward")
	}
	if p.VerifiedAt.Location() != time.UTC {
		t.Fatalf("expected VerifiedAt in UTC")
	}
}

func TestProfile_SetStatus_NonVerifiedKeepsVerifiedAtNil(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	p, err := NewProfessional(NewProfessionalParams{
		UserID:             userID,
		Kind:               KindDoctor,
		RegistrationNumber: "CRM-123",
		RegistrationIssuer: "CRM",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if p.VerifiedAt != nil {
		t.Fatalf("expected VerifiedAt nil at start")
	}
	before := p.UpdatedAt
	time.Sleep(2 * time.Millisecond)

	p.Reject()
	if p.VerifiedAt != nil {
		t.Fatalf("expected VerifiedAt to remain nil when rejected")
	}
	if !p.UpdatedAt.After(before) {
		t.Fatalf("expected UpdatedAt to move forward")
	}
}

func TestProfile_SetStatus_LeavingVerifiedPreservesVerifiedAt(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	p, err := NewProfessional(NewProfessionalParams{
		UserID:             userID,
		Kind:               KindDoctor,
		RegistrationNumber: "CRM-123",
		RegistrationIssuer: "CRM",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	p.Verify()
	if p.VerifiedAt == nil {
		t.Fatalf("expected VerifiedAt to be set")
	}
	verifiedAt := *p.VerifiedAt

	p.Reject()
	if p.Status != StatusRejected {
		t.Fatalf("expected status rejected")
	}
	if p.VerifiedAt == nil {
		t.Fatalf("expected VerifiedAt to be preserved when leaving verified")
	}
	if !p.VerifiedAt.Equal(verifiedAt) {
		t.Fatalf("expected VerifiedAt to remain the same when leaving verified")
	}
}
