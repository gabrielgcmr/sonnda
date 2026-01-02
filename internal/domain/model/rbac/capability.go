package rbac

type CapabilityLevel string

const (
	CapabilityClinical  CapabilityLevel = "clinical"   // Médicos, enfermeiros e outros profissionais de saúde
	CapabilityBasicCare CapabilityLevel = "basic_care" // Pacientes e cuidadores
	CapabilityAdmin     CapabilityLevel = "admin"      // Administradores do sistema
)

func (cl CapabilityLevel) IsValid() bool {
	switch cl {
	case CapabilityClinical, CapabilityBasicCare, CapabilityAdmin:
		return true
	default:
		return false
	}
}

func CapabilityForRole(role Role) CapabilityLevel {
	switch role {
	case RoleAdmin:
		return CapabilityAdmin
	case RoleDoctor, RoleNurse:
		return CapabilityClinical
	case RolePatient, RoleCaregiver:
		return CapabilityBasicCare
	default:
		return ""
	}
}
