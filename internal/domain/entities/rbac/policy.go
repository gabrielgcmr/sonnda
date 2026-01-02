package rbac

type PolicyService struct{}

func NewPolicyService() *PolicyService {
	return &PolicyService{}
}

func (ps *PolicyService) CanPreform(role Role, action Action) bool {

	switch action {
	//Dados cadastrais do paciente
	case ActionReadPatientDemographics:
		return role == RoleDoctor || role == RoleNurse || role == RoleAdmin
	case ActionUpdatePatientDemographics:
		return role == RoleDoctor || role == RoleAdmin

	default:
		return false
	}

}
