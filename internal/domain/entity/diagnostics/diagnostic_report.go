package diagnostics

import (
	"time"

	"github.com/google/uuid"
)

type DiagnosticReport struct {
	ID        uuid.UUID `json:"id"`
	PatientID uuid.UUID `json:"patient_id"`

	// Status é vital no FHIR: preliminary, final, amended, corrected
	Status string `json:"status"`

	// Category: No seu caso, "LAB" (Laboratory)
	Category string `json:"category"`

	// O código do relatório (ex: LOINC do painel completo)
	Code string `json:"code"`

	// Metadados do emissor (Laboratório)
	PerformerName string `json:"performer_name"`

	// Datas clínicas
	EffectiveDateTime time.Time `json:"effective_datetime"` // Quando a amostra foi colhida
	Issued            time.Time `json:"issued"`             // Quando o laudo foi liberado

	// No FHIR, o relatório contém referências para Observações
	Observations []Observation `json:"observations"`

	// Campo para IA salvar o texto bruto se necessário
	Conclusion *string `json:"conclusion,omitempty"`
}
