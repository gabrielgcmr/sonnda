package rbac

import "github.com/gabrielgcmr/sonnda/internal/domain/model/user"

type CapabilityLevel string

const (
	CapabilityClinical  CapabilityLevel = "clinical"   // Médicos, enfermeiros e outros profissionais de saúde
	CapabilityBasicCare CapabilityLevel = "basic_care" // Pacientes e cuidadores
	//CapabilityAdmin     CapabilityLevel = "admin"      // Administradores do sistema
)

func (cl CapabilityLevel) IsValid() bool {
	switch cl {
	case CapabilityClinical, CapabilityBasicCare:
		return true
	default:
		return false
	}
}

func CapabilityForAccountType(at user.AccountType) CapabilityLevel {
	switch at {
	case user.AccountTypeProfessional:
		return CapabilityClinical
	case user.AccountTypeBasicCare:
		return CapabilityBasicCare
	default:
		return ""
	}
}
