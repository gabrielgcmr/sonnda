// internal/domain/entity/medicalrecord/problem/problem.go
package problem

import "github.com/google/uuid"

type Problem struct {
	ID              uuid.UUID
	MedicalRecordID uuid.UUID
	Name            string
	Abbreviation    string
	BodySystem      string
	Description     string
	Other           string
}
