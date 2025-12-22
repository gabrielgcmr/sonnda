package medicalrecord

import "time"

type MedicalRecord struct {
	ID          string
	PatientID   string
	CreatedBy   string
	EntryType   MedicalRecordType
	Title       string
	Description string
	Date        time.Time
	CreatedAt   time.Time
}

type MedicalRecordType string

const (
	RecordTypePrevention   MedicalRecordType = "PREVENTION"
	RecordTypeProblem      MedicalRecordType = "PROBLEM"
	RecordTypeLabsExam     MedicalRecordType = "LABS_EXAM"
	RecordTypeImageExam    MedicalRecordType = "IMAGE_EXAM"
	RecordTypePhysicalExam MedicalRecordType = "PHYSICAL_EXAM"
	RecordTypeNote         MedicalRecordType = "NOTE"
)
