package rbac

type PolicyService struct{}

func NewPolicyService() *PolicyService {
	return &PolicyService{}
}

func (ps *PolicyService) CanPreform(role Role, action Action) bool {
	// No MVP, administradores e médicos têm permissão para todas as ações
	//Depois avaliar quais as capabilitis do médico
	if role == RoleAdmin || role == RoleDoctor {
		return true
	}

	level := CapabilityForRole(role)
	switch action {
	case ActionRecordMeasurement:
		return level == CapabilityBasicCare || level == CapabilityClinical

	case ActionWriteClinicalNote:
		return level == CapabilityClinical

	default:
		return false
	}

}
