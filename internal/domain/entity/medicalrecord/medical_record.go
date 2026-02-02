// internal/domain/entity/medicalrecord/medical_record.go
package medicalrecord

import (
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/medicalrecord/antecedents"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/medicalrecord/physical"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/medicalrecord/prevention"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/medicalrecord/problem"

	"github.com/google/uuid"
)

// MedicalRecord representa o prontuário do paciente (agregado).
// A ideia é concentrar aqui as seções clínicas (antecedentes, problemas, prevenções e exames).
// Persistência/consultas podem continuar sendo feitas por tabelas especializadas (ex.: lab_reports)
// e compostas em nível de use case.
type MedicalRecord struct {
	ID        uuid.UUID
	PatientID uuid.UUID

	Antecedents antecedents.Antecedents

	Problems      []problem.Problem
	Preventions   []prevention.Prevention
	PhysicalExams []physical.PhysicalExam
	LabReports    []labs.LabReport

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Entry representa um item de timeline do prontuário.
// (Mantém o modelo antigo para não perder a ideia de "entradas" no histórico.)
type Entry struct {
	ID        uuid.UUID
	PatientID uuid.UUID
	CreatedBy uuid.UUID
	Type      EntryType
	Title     string
	Body      string
	Date      time.Time
	CreatedAt time.Time
}

type EntryType string

const (
	EntryTypeAllergy      EntryType = "ALLERGY"
	EntryTypeMedication   EntryType = "MEDICATION"
	EntryTypePrevention   EntryType = "PREVENTION"
	EntryTypeProblem      EntryType = "PROBLEM"
	EntryTypeLabExam      EntryType = "LAB_EXAM"
	EntryTypeImageExam    EntryType = "IMAGE_EXAM"
	EntryTypePhysicalExam EntryType = "PHYSICAL_EXAM"
	EntryTypeNote         EntryType = "NOTE"
)
