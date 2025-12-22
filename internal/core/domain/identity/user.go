package identity

import "time"

type Role string

const (
	RolePatient   Role = "patient"
	RoleDoctor    Role = "doctor"
	RoleCaregiver Role = "caregiver"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID           string
	AuthProvider string
	AuthSubject  string
	Email        string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) IsDoctor() bool { return u.Role == RoleDoctor }
func (u *User) IsAdmin() bool  { return u.Role == RoleAdmin }

func NewUser(
	authProvider, authSubject, email string,
	role Role,
) (*User, error) {
	if authProvider == "" {
		return nil, ErrInvalidAuthProvider
	}
	if authSubject == "" {
		return nil, ErrInvalidAuthSubject
	}
	if email == "" {
		return nil, ErrInvalidEmail
	}
	if role != RoleAdmin && role != RoleDoctor && role != RolePatient {
		return nil, ErrInvalidRole
	}

	now := time.Now()

	return &User{
		AuthProvider: authProvider,
		AuthSubject:  authSubject,
		Email:        email,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
