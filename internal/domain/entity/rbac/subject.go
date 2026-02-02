// internal/domain/entity/rbac/subject.go
package rbac

import (
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/professional"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/user"
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
