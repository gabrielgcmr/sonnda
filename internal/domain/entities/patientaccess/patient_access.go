package patientaccess

import (
	"time"

	"github.com/google/uuid"
)

// PatientAccess representa o vínculo (membership) entre um app user e um paciente.
// É intencionalmente separado do agregado Patient (relação N:N, regras próprias).
//
// Modelo alinhado à tabela patient_access existente (patient_id, user_id, role, permissions, timestamps).
// Regras como "link por CPF+birthDate" pertencem ao usecase (envolvem lookup em repositório).
//
// Sem aprovação: o vínculo pode ser criado diretamente como ativo (não modelamos status aqui).
// Revogação pode ser modelada por remoção do registro ou por um status futuramente.
//
// Campos privados para manter invariantes do domínio.
// IDs são strings por ora para evitar acoplamento com infraestrutura.
//
// O papel aqui é no contexto do paciente (membership), não o papel global do usuário.

type MemberRole string

const (
	RoleCaregiver    MemberRole = "caregiver"
	RoleProfessional MemberRole = "professional"
)

type PatientAccess struct {
	PatientID uuid.UUID
	UserID    uuid.UUID
	Role      MemberRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPatientAccess(patientID, userID uuid.UUID, role MemberRole) (*PatientAccess, error) {

	if role == "" {
		return nil, ErrInvalidRole
	}
	if !roleIsValid(role) {
		return nil, ErrInvalidRole
	}

	now := time.Now().UTC()
	return &PatientAccess{
		PatientID: patientID,
		UserID:    userID,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (a *PatientAccess) Permissions() []Permission {
	if a == nil {
		return nil
	}
	return append([]Permission(nil), defaultPermissionsForRole(a.Role)...)
}

func (a *PatientAccess) HasPermission(perm Permission) bool {
	if a == nil || perm == "" {
		return false
	}
	for _, p := range defaultPermissionsForRole(a.Role) {
		if p == perm {
			return true
		}
	}
	return false
}

func (a *PatientAccess) SetRole(role MemberRole) error {
	if a == nil {
		return ErrInvalidRole
	}
	if role == "" || !roleIsValid(role) {
		return ErrInvalidRole
	}
	a.Role = role
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func roleIsValid(role MemberRole) bool {
	switch role {
	case RoleCaregiver, RoleProfessional:
		return true
	default:
		return false
	}
}
