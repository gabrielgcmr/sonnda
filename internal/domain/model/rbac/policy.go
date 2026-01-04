package rbac

import (
	"sonnda-api/internal/domain/model/user/professional"
)

type RbacPolicy struct{}

func NewRbacPolicy() *RbacPolicy {
	return &RbacPolicy{}
}

func (ps *RbacPolicy) CanPerform(subject Subject, action Action) bool {
	subject = subject.Normalize()

	level := CapabilityForAccountType(subject.AccountType)
	if level == "" {
		return false
	}

	if level == CapabilityAdmin {
		return true
	}

	isProfessional := level == CapabilityClinical
	isBasicCare := level == CapabilityBasicCare

	switch action {
	// Patient
	case ActionReadPatient:
		return isProfessional || isBasicCare
	case ActionCreatePatient, ActionUpdatePatient:
		return isProfessional
	case ActionSoftDeletePatient:
		return false

	// Clinical
	case ActionRecordMeasurement:
		return isProfessional || isBasicCare
	case ActionWriteClinicalNote:
		return isProfessional

	// Labs
	case ActionReadLabs:
		return isProfessional || isBasicCare
	case ActionUploadLabs:
		return isProfessional || isBasicCare

	// Prescriptions
	case ActionReadPrescriptions:
		return isProfessional || isBasicCare
	case ActionWritePrescriptions:
		if !isProfessional || subject.ProfessionalKind == nil {
			return false
		}
		return *subject.ProfessionalKind == professional.KindDoctor

	default:
		return false
	}
}
