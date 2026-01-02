package prevention

import "github.com/google/uuid"

type Prevention struct {
	ID              uuid.UUID
	MedicalRecordID uuid.UUID
	Name            string
	Abbreviation    string
	Value           string
	ReferenceValue  string
	Unit            string
	Classification  string
	Description     string
	Other           string
}
