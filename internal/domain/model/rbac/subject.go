package rbac

import (
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
	"github.com/gabrielgcmr/sonnda/internal/domain/model/professional"
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
