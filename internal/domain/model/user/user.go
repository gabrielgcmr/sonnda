package user

import (
	"sonnda-api/internal/domain/model/rbac"
	"sonnda-api/internal/domain/model/shared"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	AuthProvider string
	AuthSubject  string
	Email        string
	FullName     string
	Role         rbac.Role
	BirthDate    time.Time
	CPF          string
	Phone        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// 1. Crie uma struct para agrupar os parâmetros
type NewUserParams struct {
	AuthProvider string
	AuthSubject  string
	Email        string
	FullName     string
	Role         rbac.Role
	BirthDate    time.Time
	CPF          string
	Phone        string
}

// 2. Crie um método que sabe se limpar
// Como é um ponteiro receiver (*NewUserParams), ele altera os dados originais
func (p *NewUserParams) Normalize() {
	p.AuthProvider = strings.TrimSpace(p.AuthProvider)
	p.AuthSubject = strings.TrimSpace(p.AuthSubject)
	p.Email = strings.ToLower(strings.TrimSpace(p.Email))
	p.FullName = strings.TrimSpace(p.FullName)
	p.Phone = strings.TrimSpace(p.Phone)

	// Normalize específica
	p.CPF = shared.CleanDigits(p.CPF)
}

func NewUser(params NewUserParams) (*User, error) {
	// 1. Normalize (Limpeza)
	// Chamamos funções auxiliares para não poluir o construtor
	params.Normalize()

	// 2. Criação da entidade
	now := time.Now().UTC()
	u := &User{
		ID:           uuid.Must(uuid.NewV7()),
		AuthProvider: params.AuthProvider,
		AuthSubject:  params.AuthSubject,
		Email:        params.Email,
		FullName:     params.FullName,
		Role:         params.Role,
		BirthDate:    params.BirthDate.UTC(),
		CPF:          params.CPF,
		Phone:        params.Phone,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 2. Validação
	// O construtor delega a validação para o método da própria struct
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
	if u.BirthDate.IsZero() {
		return ErrInvalidBirthDate
	}
	if u.CPF == "" || len(u.CPF) != 11 {
		return ErrInvalidCPF
	}
	// Futuramente: if len(u.CPF) != 11 { ... }

	return nil
}
