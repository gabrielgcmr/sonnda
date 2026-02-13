package diagnostics

import "github.com/google/uuid"

// Observation representa cada resultado individual (FHIR Observation)
type Observation struct {
	ID       uuid.UUID `json:"id"`
	ReportID uuid.UUID `json:"report_id"`

	Status string `json:"status"` // final, preliminary

	// Code: O coração da interoperabilidade (LOINC Code)
	// Ex: 718-7 para Hemoglobina
	Code    string `json:"code"`
	Display string `json:"display"` // Nome legível: "Hemoglobina"

	// Value: FHIR separa valor, unidade e sistema
	ValueQuantity ObservationValue `json:"value_quantity"`

	// Interpretation: H (High), L (Low), N (Normal)
	Interpretation string `json:"interpretation,omitempty"`

	// ReferenceRange: Onde a IA salva os valores de referência extraídos
	ReferenceRange string `json:"reference_range,omitempty"`
}

type ObservationValue struct {
	Value  string `json:"value"`  // Mantemos string para aceitar "< 0.01" ou "Positivo"
	Unit   string `json:"unit"`   // Ex: "mg/dL"
	System string `json:"system"` // Geralmente "http://unitsofmeasure.org" (UCUM)
}
