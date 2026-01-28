package user

import (
	"errors"
	"testing"
	"time"
)

func TestNewUser_Success_NormalizesAndSetsUTC(t *testing.T) {
	birthDate := time.Now().Add(-24 * time.Hour)

	u, err := NewUser(NewUserParams{
		AuthIssuer:  " firebase ",
		AuthSubject: " sub-123 ",
		Email:       " Person@Example.COM ",
		AccountType: AccountTypeProfessional,
		FullName:    "  Pessoa Teste  ",
		BirthDate:   birthDate,
		CPF:         "529.982.247-25",
		Phone:       "  11999999999  ",
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if u == nil {
		t.Fatalf("expected user, got nil")
	}

	if u.AuthIssuer != "firebase" {
		t.Errorf("expected auth provider trimmed, got '%s'", u.AuthIssuer)
	}
	if u.AuthSubject != "sub-123" {
		t.Errorf("expected auth subject trimmed, got '%s'", u.AuthSubject)
	}
	if u.Email != "person@example.com" {
		t.Errorf("expected email trimmed, got '%s'", u.Email)
	}
	if u.FullName != "Pessoa Teste" {
		t.Errorf("expected full name trimmed, got '%s'", u.FullName)
	}
	if u.AccountType != AccountTypeProfessional {
		t.Errorf("expected professional account type")
	}
	if u.Phone != "11999999999" {
		t.Errorf("expected phone trimmed, got '%s'", u.Phone)
	}

	if u.CPF != "52998224725" {
		t.Errorf("expected cpf cleaned (digits only), got '%s'", u.CPF)
	}

	if u.CreatedAt.Location() != time.UTC {
		t.Errorf("expected CreatedAt in UTC")
	}
	if u.UpdatedAt.Location() != time.UTC {
		t.Errorf("expected UpdatedAt in UTC")
	}
	if u.BirthDate.Location() != time.UTC {
		t.Errorf("expected BirthDate in UTC")
	}
}

func TestNewUser_ValidationErrors(t *testing.T) {
	validDate := time.Now().Add(-24 * time.Hour)

	cases := []struct {
		name string
		fn   func() NewUserParams
		err  error
	}{
		{
			name: "missing authProvider",
			err:  ErrInvalidAuthIssuer,
			fn: func() NewUserParams {
				params := validParams(validDate)
				params.AuthIssuer = "   "
				return params
			},
		},
		{
			name: "missing authSubject",
			err:  ErrInvalidAuthSubject,
			fn: func() NewUserParams {
				params := validParams(validDate)
				params.AuthSubject = "   "
				return params
			},
		},
		{
			name: "missing email",
			err:  ErrInvalidEmail,
			fn: func() NewUserParams {
				params := validParams(validDate)
				params.Email = "   "
				return params
			},
		},
		{
			name: "missing fullName",
			err:  ErrInvalidFullName,
			fn: func() NewUserParams {
				params := validParams(validDate)
				params.FullName = "   "
				return params
			},
		},
		{
			name: "invalid accountType",
			err:  ErrInvalidAccountType,
			fn: func() NewUserParams {
				params := validParams(validDate)
				params.AccountType = AccountType("nope")
				return params
			},
		},
		{
			name: "birthDate zero",
			err:  ErrInvalidBirthDate,
			fn: func() NewUserParams {
				params := validParams(validDate)
				params.BirthDate = time.Time{}
				return params
			},
		},
		{
			name: "missing cpf",
			err:  ErrInvalidCPF,
			fn: func() NewUserParams {
				params := validParams(validDate)
				params.CPF = "   "
				return params
			},
		},
		{
			name: "missing phone",
			err:  ErrInvalidPhone,
			fn: func() NewUserParams {
				params := validParams(validDate)
				params.Phone = "   "
				return params
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewUser(tc.fn())
			if err == nil {
				t.Fatalf("expected error '%v', got nil", tc.err)
			}
			if !errors.Is(err, tc.err) {
				t.Errorf("expected error '%v', got '%v'", tc.err, err)
			}
		})
	}
}

func validParams(birthDate time.Time) NewUserParams {
	return NewUserParams{
		AuthIssuer:  "firebase",
		AuthSubject: "sub",
		Email:       "a@b.com",
		FullName:    "User",
		AccountType: AccountTypeProfessional,
		BirthDate:   birthDate,
		CPF:         "12345678901",
		Phone:       "11",
	}
}

func TestUser_ApplyUpdate_Idempotent(t *testing.T) {
	birthDate := time.Now().Add(-24 * time.Hour)

	u, err := NewUser(NewUserParams{
		AuthIssuer:  "firebase",
		AuthSubject: "sub-123",
		Email:       "person@example.com",
		AccountType: AccountTypeProfessional,
		FullName:    "Pessoa Teste",
		BirthDate:   birthDate,
		CPF:         "529.982.247-25",
		Phone:       "  11999999999  ",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	before := u.UpdatedAt

	changed, err := u.ApplyUpdate(UpdateUserParams{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if changed {
		t.Fatalf("expected changed=false")
	}
	if !u.UpdatedAt.Equal(before) {
		t.Fatalf("expected UpdatedAt unchanged")
	}

	sameName := "  Pessoa Teste  "
	changed, err = u.ApplyUpdate(UpdateUserParams{FullName: &sameName})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if changed {
		t.Fatalf("expected changed=false for no-op update")
	}
	if !u.UpdatedAt.Equal(before) {
		t.Fatalf("expected UpdatedAt unchanged on no-op update")
	}

	sameCPF := "529.982.247-25"
	changed, err = u.ApplyUpdate(UpdateUserParams{CPF: &sameCPF})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if changed {
		t.Fatalf("expected changed=false for equivalent CPF update")
	}

	samePhone := "11999999999"
	changed, err = u.ApplyUpdate(UpdateUserParams{Phone: &samePhone})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if changed {
		t.Fatalf("expected changed=false for equivalent phone update")
	}

	brt := time.FixedZone("BRT", -3*3600)
	sameBirth := u.BirthDate.In(brt)
	changed, err = u.ApplyUpdate(UpdateUserParams{BirthDate: &sameBirth})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if changed {
		t.Fatalf("expected changed=false for equivalent birthDate update")
	}

	time.Sleep(2 * time.Millisecond)
	newName := "Novo Nome"
	changed, err = u.ApplyUpdate(UpdateUserParams{FullName: &newName})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true")
	}
	if u.FullName != "Novo Nome" {
		t.Fatalf("expected FullName to be updated")
	}
	if !u.UpdatedAt.After(before) {
		t.Fatalf("expected UpdatedAt to move forward")
	}
}

func TestUser_ApplyUpdate_RejectsInvalidAndDoesNotMutate(t *testing.T) {
	birthDate := time.Now().Add(-24 * time.Hour)

	u, err := NewUser(validParams(birthDate))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	before := *u
	invalidPhone := "   "
	newName := "Nome Novo"

	_, err = u.ApplyUpdate(UpdateUserParams{
		FullName: &newName,
		Phone:    &invalidPhone,
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("expected %v, got %v", ErrInvalidPhone, err)
	}
	if u.FullName != before.FullName {
		t.Fatalf("expected FullName unchanged on error")
	}
	if !u.UpdatedAt.Equal(before.UpdatedAt) {
		t.Fatalf("expected UpdatedAt unchanged on error")
	}
}
