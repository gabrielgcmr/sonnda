// internal/domain/entity/professional/kind.go
package professional

import "strings"

type Kind string

const (
	//Ativo no mvp
	KindDoctor Kind = "doctor" // Médico (CRM)
	//Fora do MVP
	KindNurse           Kind = "nurse"           // Enfermeiro(a) (COREN)
	KindNursingTech     Kind = "nursing_tech"    // Técnico(a) de Enfermagem
	KindPhysiotherapist Kind = "physiotherapist" // Fisioterapeuta (CREFITO)
	KindPsychologist    Kind = "psychologist"    // Psicólogo(a) (CRP)
	KindNutritionist    Kind = "nutritionist"    // Nutricionista (CRN)
	KindPharmacist      Kind = "pharmacist"      // Farmacêutico(a) (CRF)
	KindDentist         Kind = "dentist"         // Cirurgião-Dentista (CRO)
)

func (k Kind) IsValid() bool {
	switch k {
	case
		KindDoctor,
		KindNurse,
		KindNursingTech,
		KindPhysiotherapist,
		KindPsychologist,
		KindNutritionist,
		KindPharmacist,
		KindDentist:
		return true
	default:
		return false
	}
}

func (k Kind) Normalize() Kind {
	return Kind(strings.ToLower(strings.TrimSpace(string(k))))
}
