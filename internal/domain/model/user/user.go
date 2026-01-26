// internal/domain/model/user/user.go
package user

import (
	"fmt"
	"strings"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/model/demographics"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	AuthProvider string    `json:"auth_provider"`
	AuthSubject  string    `json:"auth_subject"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	AccountType  AccountType `json:"account_type"`
	BirthDate    time.Time `json:"birth_date"`
	CPF          string    `json:"cpf"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpdateUserParams struct {
	FullName  *string
	BirthDate *time.Time
	CPF       *string
	Phone     *string
}

// 1. Crie uma struct para agrupar os parâmetros.
type NewUserParams struct {
	AuthProvider string
	AuthSubject  string
	Email        string
	FullName     string
	AccountType  AccountType
	BirthDate    time.Time
	CPF          string
	Phone        string
}

// 2. Crie um método que sabe se limpar.
// Como é um ponteiro receiver (*NewUserParams), ele altera os dados originais.
func (p *NewUserParams) Normalize() {
	p.AuthProvider = strings.TrimSpace(p.AuthProvider)
	p.AuthSubject = strings.TrimSpace(p.AuthSubject)
	p.Email = strings.ToLower(strings.TrimSpace(p.Email))
	p.FullName = strings.TrimSpace(p.FullName)
	p.Phone = strings.TrimSpace(p.Phone)

	p.AccountType = p.AccountType.Normalize()
	p.CPF = demographics.CleanDigits(p.CPF)
}

func NewUser(params NewUserParams) (*User, error) {
	params.Normalize()

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}

	now := time.Now().UTC()
	u := &User{
		ID:           id,
		AuthProvider: params.AuthProvider,
		AuthSubject:  params.AuthSubject,
		Email:        params.Email,
		FullName:     params.FullName,
		AccountType:  params.AccountType,
		BirthDate:    params.BirthDate.UTC(),
		CPF:          params.CPF,
		Phone:        params.Phone,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := u.Validate(); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) Validate() error {
	if u.AuthProvider == "" {
		return ErrInvalidAuthProvider
	}
	if u.AuthSubject == "" {
		return ErrInvalidAuthSubject
	}
	if u.Email == "" {
		return ErrInvalidEmail
	}
	if u.FullName == "" {
		return ErrInvalidFullName
	}
	if !u.AccountType.IsValid() {
		return ErrInvalidAccountType
	}
	if u.BirthDate.IsZero() {
		return ErrInvalidBirthDate
	}
	if u.CPF == "" || len(u.CPF) != 11 {
		return ErrInvalidCPF
	}
	if u.Phone == "" {
		return ErrInvalidPhone
	}

	return nil
}

// ApplyUpdate aplica atualizações com idempotência:
// - Se nenhum campo mudar, UpdatedAt não é alterado.
// - Se algum campo mudar, UpdatedAt é atualizado para UTC.
func (u *User) ApplyUpdate(params UpdateUserParams) (changed bool, err error) {
	if u == nil {
		return false, fmt.Errorf("user is nil")
	}

	nextFullName := u.FullName
	nextBirthDate := u.BirthDate
	nextCPF := u.CPF
	nextPhone := u.Phone

	if params.FullName != nil {
		name := strings.TrimSpace(*params.FullName)
		if name == "" {
			return false, ErrInvalidFullName
		}
		nextFullName = name
	}

	if params.BirthDate != nil {
		birthDate := params.BirthDate.UTC()
		if birthDate.IsZero() || birthDate.After(time.Now().UTC()) {
			return false, ErrInvalidBirthDate
		}
		nextBirthDate = birthDate
	}

	if params.CPF != nil {
		cpf := demographics.CleanDigits(*params.CPF)
		if cpf == "" || len(cpf) != 11 {
			return false, ErrInvalidCPF
		}
		nextCPF = cpf
	}

	if params.Phone != nil {
		phone := strings.TrimSpace(*params.Phone)
		if phone == "" {
			return false, ErrInvalidPhone
		}
		nextPhone = phone
	}

	changed = nextFullName != u.FullName ||
		!nextBirthDate.Equal(u.BirthDate) ||
		nextCPF != u.CPF ||
		nextPhone != u.Phone

	if !changed {
		return false, nil
	}

	u.FullName = nextFullName
	u.BirthDate = nextBirthDate
	u.CPF = nextCPF
	u.Phone = nextPhone
	u.UpdatedAt = time.Now().UTC()

	return true, nil
}
