package patientaccess

type RelationshipType string //O relacionamento que o usuário tem com o paciente

const (
	RelationshipTypeCaregiver    RelationshipType = "caregiver"
	RelationshipTypeFamily       RelationshipType = "family"
	RelationshipTypeProfessional RelationshipType = "professional"
	RelationshipTypeSelf         RelationshipType = "self" // Quando o proprio paciente é também o usuário
)

func (rt RelationshipType) IsValid() bool {
	switch rt {
	case RelationshipTypeCaregiver,
		RelationshipTypeFamily,
		RelationshipTypeProfessional,
		RelationshipTypeSelf:
		return true
	default:
		return false
	}
}
