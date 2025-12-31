package patientaccess

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewPatientAccess_Success_DefaultsAndUTC(t *testing.T) {
	patientID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	a, err := NewPatientAccess(patientID, userID, RoleCaregiver)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if a == nil {
		t.Fatalf("expected access, got nil")
	}
	if a.PatientID != patientID {
		t.Fatalf("expected patient id, got %q", a.PatientID)
	}
	if a.UserID != userID {
		t.Fatalf("expected user id, got %q", a.UserID)
	}
	if a.Role != RoleCaregiver {
		t.Fatalf("expected role caregiver")
	}
	if a.CreatedAt.Location() != time.UTC {
		t.Fatalf("expected CreatedAt in UTC")
	}
	if a.UpdatedAt.Location() != time.UTC {
		t.Fatalf("expected UpdatedAt in UTC")
	}
}

func TestNewPatientAccess_InvalidInputs(t *testing.T) {
	cases := []struct {
		name string
		err  error
		fn   func() error
	}{
		{
			name: "missing role",
			err:  ErrInvalidRole,
			fn: func() error {
				patientID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
				userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
				_, err := NewPatientAccess(patientID, userID, "")
				return err
			},
		},
		{
			name: "invalid role",
			err:  ErrInvalidRole,
			fn: func() error {
				patientID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
				userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
				_, err := NewPatientAccess(patientID, userID, MemberRole("banana"))
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

func TestPatientAccess_Permissions_ReturnsCopy(t *testing.T) {
	patientID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	a, err := NewPatientAccess(patientID, userID, RoleCaregiver)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	p1 := a.Permissions()
	p2 := a.Permissions()
	if len(p1) == 0 {
		t.Fatalf("expected some permissions")
	}
	if len(p1) != len(p2) {
		t.Fatalf("expected same length")
	}

	p1[0] = Permission("tamper")
	p3 := a.Permissions()
	if p3[0] == Permission("tamper") {
		t.Fatalf("expected permissions to be a defensive copy")
	}
}

func TestPatientAccess_HasPermission(t *testing.T) {
	patientID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	a, err := NewPatientAccess(patientID, userID, RoleCaregiver)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if a.HasPermission("") {
		t.Fatalf("expected empty permission to be false")
	}
	if !a.HasPermission(PermPatientRead) {
		t.Fatalf("expected to have patient:read")
	}
	if a.HasPermission(Permission("banana")) {
		t.Fatalf("expected unknown permission to be false")
	}
}

func TestPatientAccess_HasPermission_ProfessionalBypass(t *testing.T) {
	patientID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	a, err := NewPatientAccess(patientID, userID, RoleProfessional)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !a.HasPermission(PermMedicalProblemWrite) {
		t.Fatalf("expected professional to have medical_record:problem:write")
	}
	if !a.HasPermission(PermMedicalPrevention) {
		t.Fatalf("expected professional to have medical_record:prevention:write")
	}
	if !a.HasPermission(PermLabsUpload) {
		t.Fatalf("expected professional to have medical_record:labs:upload")
	}
	if a.HasPermission(Permission("banana")) {
		t.Fatalf("expected unknown permission to be false")
	}
}

func TestPatientAccess_SetRole_ValidatesAndUpdatesTimestamp(t *testing.T) {
	patientID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	a, err := NewPatientAccess(patientID, userID, RoleCaregiver)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := a.SetRole(MemberRole("banana")); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}

	before := a.UpdatedAt
	time.Sleep(2 * time.Millisecond)

	if err := a.SetRole(RoleProfessional); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if a.Role != RoleProfessional {
		t.Fatalf("expected role professional")
	}
	if !a.UpdatedAt.After(before) {
		t.Fatalf("expected UpdatedAt to move forward")
	}
	if a.UpdatedAt.Location() != time.UTC {
		t.Fatalf("expected UpdatedAt in UTC")
	}
}
