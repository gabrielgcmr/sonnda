// internal/features/patient/access/dto.go
package patientaccess

import "github.com/google/uuid"

// PatientSummary is the public patient data returned by access listings.
// Relationship metadata remains internal until the domain defines its meaning.
type PatientSummary struct {
	ID        uuid.UUID
	FullName  string
	AvatarURL *string
}

type ListPatientsOutput struct {
	Patients []PatientSummary
	Total    int64
	Limit    int
	Offset   int
}
