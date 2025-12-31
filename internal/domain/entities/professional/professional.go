package professional

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type VerificationStatus string

const (
	StatusPending  VerificationStatus = "pending"
	StatusVerified VerificationStatus = "verified"
	StatusRejected VerificationStatus = "rejected"
)

func (s VerificationStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusVerified, StatusRejected:
		return true
	default:
		return false
	}
}

// 1. Parameter Object (Renomeado)
type NewProfessionalParams struct {
	UserID             uuid.UUID
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
}

// 2. Normalize (Limpeza)
func (p *NewProfessionalParams) Normalize() {
	p.RegistrationNumber = strings.TrimSpace(p.RegistrationNumber)
	p.RegistrationIssuer = strings.TrimSpace(p.RegistrationIssuer)

	if p.RegistrationState != nil {
		trimmed := strings.TrimSpace(*p.RegistrationState)
		p.RegistrationState = &trimmed
	}
}

// 3. Entidade Renomeada: De Profile para Professional
// Representa o ATOR no sistema, não apenas um "perfil".
type Professional struct {
	UserID             uuid.UUID
	RegistrationNumber string
	RegistrationIssuer string
	RegistrationState  *string
	Status             VerificationStatus
	VerifiedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// 4. Construtor (Factory)
// Uso: professional.NewProfessional(...)
func NewProfessional(params NewProfessionalParams) (*Professional, error) {
	// Limpa
	params.Normalize()

	// Valida (Check rápido antes de criar struct)
	// Como movemos a validação para dentro do método Validate(),
	// podemos instanciar primeiro ou checar aqui os params.
	// Vamos criar a struct primeiro para usar o método Validate().

	now := time.Now().UTC()

	// MVP: Como você disse que vai testar com amigos, você decide se nasce
	// como "Pending" (pra você aprovar no banco) ou "Verified" direto.
	// Mantive Pending por segurança.
	prof := &Professional{
		UserID:             params.UserID,
		RegistrationNumber: params.RegistrationNumber,
		RegistrationIssuer: params.RegistrationIssuer,
		RegistrationState:  params.RegistrationState,
		Status:             StatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := prof.Validate(); err != nil {
		return nil, err
	}

	return prof, nil
}

func (p *Professional) Validate() error {
	if p.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if p.RegistrationNumber == "" {
		return ErrInvalidRegistrationNumber
	}
	if p.RegistrationIssuer == "" {
		return ErrInvalidRegistrationIssuer
	}
	return nil
}

// Métodos de domínio
func (p *Professional) Verify() {
	now := time.Now().UTC()
	p.Status = StatusVerified
	p.VerifiedAt = &now
	p.UpdatedAt = now
}

func (p *Professional) Reject() {
	now := time.Now().UTC()
	p.Status = StatusRejected
	p.UpdatedAt = now
}
