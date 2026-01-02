package patient

import (
	"errors"
	"testing"
	"time"

	"sonnda-api/internal/domain/model/shared"

	"github.com/google/uuid"
)

func TestNewPatient_Success_NormalizesAndSetsUTC(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cns := " 123456789012345 "
	phone := " 11999999999 "
	birthDate := time.Now().Add(-24 * time.Hour)

	p, err := NewPatient(NewPatientParams{
		UserID:    &userID,
		CPF:       "52998224725",
		CNS:       &cns,
		FullName:  "  Paciente Teste  ",
		BirthDate: birthDate,
		Gender:    shared.Gender("female"),
		Race:      shared.Race("white"),
		Phone:     &phone,
		AvatarURL: "  https://example.com/a.png  ",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if p == nil {
		t.Fatalf("expected patient, got nil")
	}
	if p.FullName != "Paciente Teste" {
		t.Fatalf("expected full name trimmed")
	}
	if p.AvatarURL != "https://example.com/a.png" {
		t.Fatalf("expected avatar url trimmed")
	}
	if p.CreatedAt.Location() != time.UTC {
		t.Fatalf("expected CreatedAt in UTC")
	}
	if p.UpdatedAt.Location() != time.UTC {
		t.Fatalf("expected UpdatedAt in UTC")
	}
	if p.CPF == "" {
		t.Fatalf("expected CPF to be set")
	}
	if p.OwnerUserID == nil || *p.OwnerUserID != userID {
		t.Fatalf("expected UserID to be present")
	}
	if p.CNS == nil || *p.CNS != "123456789012345" {
		t.Fatalf("expected CNS to be present")
	}
	if p.Phone == nil || *p.Phone != "11999999999" {
		t.Fatalf("expected phone to be present")
	}
}

func TestNewPatient_InvalidInputs(t *testing.T) {
	now := time.Now()
	birthDate := now.Add(-24 * time.Hour)

	cases := []struct {
		name string
		err  error
		fn   func() NewPatientParams
	}{
		{
			name: "missing fullName",
			err:  ErrInvalidFullName,
			fn: func() NewPatientParams {
				params := validParams(birthDate)
				params.FullName = "   "
				return params
			},
		},
		{
			name: "invalid cpf",
			err:  shared.ErrInvalidCPF,
			fn: func() NewPatientParams {
				params := validParams(birthDate)
				params.CPF = "123"
				return params
			},
		},
		{
			name: "birthDate zero",
			err:  shared.ErrInvalidBirthDate,
			fn: func() NewPatientParams {
				params := validParams(birthDate)
				params.BirthDate = time.Time{}
				return params
			},
		},
		{
			name: "birthDate future",
			err:  shared.ErrInvalidBirthDate,
			fn: func() NewPatientParams {
				params := validParams(birthDate)
				params.BirthDate = now.Add(24 * time.Hour)
				return params
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewPatient(tc.fn())
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected %v, got %v", tc.err, err)
			}
		})
	}
}

func TestPatient_ApplyUpdate_NormalizesAndUpdatesTimestamp(t *testing.T) {
	p, err := NewPatient(NewPatientParams{
		CPF:       "52998224725",
		FullName:  "Paciente",
		BirthDate: time.Now().Add(-24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	before := p.UpdatedAt
	oldName := p.FullName

	emptyName := "   "
	newAvatar := "  http://img  "
	time.Sleep(2 * time.Millisecond)

	p.ApplyUpdate(&emptyName, nil, &newAvatar, nil, nil, nil)

	if p.FullName != oldName {
		t.Fatalf("expected empty name update to be ignored")
	}
	if p.AvatarURL != "http://img" {
		t.Fatalf("expected avatar url trimmed")
	}
	if !p.UpdatedAt.After(before) {
		t.Fatalf("expected UpdatedAt to move forward")
	}
}

func validParams(birthDate time.Time) NewPatientParams {
	return NewPatientParams{
		CPF:       "52998224725",
		FullName:  "Paciente",
		BirthDate: birthDate,
		Gender:    shared.GenderFemale,
		Race:      shared.RaceUnknown,
	}
}
