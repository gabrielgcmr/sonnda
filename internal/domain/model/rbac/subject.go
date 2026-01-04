package rbac

import (
	"sonnda-api/internal/domain/model/user"
	"sonnda-api/internal/domain/model/user/professional"
)

type Subject struct {
	AccountType      user.AccountType
	ProfessionalKind *professional.Kind
}

func (s Subject) Normalize() Subject {
	s.AccountType = s.AccountType.Normalize()
	if s.ProfessionalKind != nil {
		normalized := (*s.ProfessionalKind).Normalize()
		s.ProfessionalKind = &normalized
	}
	return s
}
