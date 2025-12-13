package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RolePatient   Role = "patient"
	RoleDoctor    Role = "doctor"
	RoleCaregiver Role = "caregiver"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	AuthProvider string    `json:"auth_provider"`
	AuthSubject  string    `json:"auth_subject"`
	Email        string    `json:"email"`
	Role         Role      `json:"role"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) IsDoctor() bool { return u.Role == RoleDoctor }
func (u *User) IsAdmin() bool  { return u.Role == RoleAdmin }

func NewUser(
	authProvider, authSubject, email string,
	role Role,
	ubsID *string,
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
