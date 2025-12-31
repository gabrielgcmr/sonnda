package medicalrecord

import (
	"time"

	"sonnda-api/internal/domain/entities/medicalrecord/antecedents"
	"sonnda-api/internal/domain/entities/medicalrecord/labs"
	"sonnda-api/internal/domain/entities/medicalrecord/physical"
	"sonnda-api/internal/domain/entities/medicalrecord/prevention"
	"sonnda-api/internal/domain/entities/medicalrecord/problem"

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
