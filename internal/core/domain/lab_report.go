// internal/core/domain/model.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

// LabReport representa o laudo inteiro (PDF/processado) ligado a um paciente.
// Uma linha em lab_reports.
type LabReport struct {
	ID        uuid.UUID `db:"id"          json:"id"`
	PatientID uuid.UUID `db:"patient_id"  json:"patient_id"`

	// Metadados extraídos do cabeçalho do laudo
	PatientName       *string    `db:"patient_name"       json:"patient_name,omitempty"`
	PatientDOB        *time.Time `db:"patient_dob"        json:"patient_dob,omitempty"`
	LabName           *string    `db:"lab_name"           json:"lab_name,omitempty"`
	LabPhone          *string    `db:"lab_phone"          json:"lab_phone,omitempty"`
	InsuranceProvider *string    `db:"insurance_provider" json:"insurance_provider,omitempty"`
	RequestingDoctor  *string    `db:"requesting_doctor"  json:"requesting_doctor,omitempty"`
	TechnicalManager  *string    `db:"technical_manager"  json:"technical_manager,omitempty"`
	ReportDate        *time.Time `db:"report_date"        json:"report_date,omitempty"`
	Fingerprint       *string    `db:"fingerprint"        json:"fingerprint,omitempty"`

	RawText *string `db:"raw_text" json:"raw_text,omitempty"`

	// Carregado via JOIN quando você quiser devolver tudo de uma vez
	TestResults []LabResult `json:"test_results,omitempty"`

	CreatedAt        time.Time `db:"created_at"         json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"         json:"updated_at"`
	UploadedByUserID uuid.UUID `db:"uploaded_by_user_id" json:"uploaded_by_user_id,omitempty"`
}

// LabResult representa um exame/painel dentro do laudo
// (ex.: Hemograma, Creatinina, HbA1c). Uma linha em lab_results.
type LabResult struct {
	ID          uuid.UUID `db:"id"            json:"id"`
	LabReportID uuid.UUID `db:"lab_report_id" json:"lab_report_id"`

	TestName string  `db:"test_name" json:"test_name"`
	Material *string `db:"material"  json:"material,omitempty"`
	Method   *string `db:"method"    json:"method,omitempty"`

	CollectedAt *time.Time `db:"collected_at" json:"collected_at,omitempty"`
	ReleaseAt   *time.Time `db:"release_at"   json:"release_at,omitempty"`

	// Carregado via JOIN quando necessário
	Items []LabResultItem `json:"items,omitempty"`
}

// LabResultItem representa uma linha/paramêtro dentro de um teste
// (ex.: Hemoglobina, Creatinina, LDL). Uma linha em lab_result_items.
type LabResultItem struct {
	ID          uuid.UUID `db:"id"                 json:"id"`
	LabResultID uuid.UUID `db:"lab_result_id" json:"lab_result_id"`

	ParameterName string  `db:"parameter_name" json:"parameter_name"`
	ResultValue   *string `db:"result_value"    json:"result_value,omitempty"`
	ResultUnit    *string `db:"result_unit"    json:"result_unit,omitempty"`
	ReferenceText *string `db:"reference_text" json:"reference_text,omitempty"`
}

// LabResultItemTimeline representa um item de exame com timestamp,
// usado para listar o histórico de um parâmetro específico.
type LabResultItemTimeline struct {
	ReportID      uuid.UUID  `db:"report_id"      json:"report_id"`
	LabResultID   uuid.UUID  `db:"lab_result_id" json:"lab_result_id"`
	ItemID        uuid.UUID  `db:"item_id"        json:"item_id"`
	ReportDate    *time.Time `db:"report_date"    json:"report_date,omitempty"`
	TestName      string     `db:"test_name"      json:"test_name"`
	ParameterName string     `db:"parameter_name" json:"parameter_name"`
	ResultValue   *string    `db:"result_value"    json:"result_value,omitempty"`
	ResultUnit    *string    `db:"result_unit"    json:"result_unit,omitempty"`
}
