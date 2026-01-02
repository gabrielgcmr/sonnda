package user

import (
	"errors"
	"sonnda-api/internal/domain/entities/rbac"
	"testing"
	"time"
)

func TestNewUser_Success_NormalizesAndSetsUTC(t *testing.T) {
	birthDate := time.Now().Add(-24 * time.Hour)

	u, err := NewUser(NewUserParams{
		AuthProvider: " firebase ",
		AuthSubject:  " sub-123 ",
		Email:        " Person@Example.COM ",
		Role:         rbac.RoleCommon,
		FullName:     "  Pessoa Teste  ",
		BirthDate:    birthDate,
		CPF:          "529.982.247-25",
		Phone:        "  11999999999  ",
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if u == nil {
		t.Fatalf("expected user, got nil")
	}

	if u.AuthProvider != "firebase" {
		t.Errorf("expected auth provider trimmed, got '%s'", u.AuthProvider)
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
	if u.Role != rbac.RoleCommon {
		t.Errorf("expected common role")
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
			err:  ErrInvalidAuthProvider,
			fn: func() NewUserParams {
				params := validParams(validDate)
				params.AuthProvider = "   "
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
		AuthProvider: "firebase",
		AuthSubject:  "sub",
		Email:        "a@b.com",
		FullName:     "User",
		Role:         rbac.RoleCommon,
		BirthDate:    birthDate,
		CPF:          "12345678901",
		Phone:        "11",
	}
}
