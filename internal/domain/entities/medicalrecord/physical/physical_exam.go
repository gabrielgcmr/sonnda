package physical

import "github.com/google/uuid"

type PhysicalExam struct {
	ID              uuid.UUID
	MedicalRecordID uuid.UUID
	SystolicBP      string
	DiastolicBP     string
}
