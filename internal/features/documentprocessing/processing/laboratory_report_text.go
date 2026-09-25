// internal/features/documentprocessing/processing/laboratory_report_text.go
package processing

import (
	"strings"
	"time"

	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
)

func buildLaboratoryReportText(report *laboratory.LabReportOutput) string {
	if report == nil {
		return ""
	}

	var builder strings.Builder
	writeLine := func(parts ...string) {
		line := strings.TrimSpace(strings.Join(parts, " "))
		if line != "" {
			builder.WriteString(line)
			builder.WriteByte('\n')
		}
	}

	writeLine("Exame laboratorial")
	writeOptionalLine(&builder, "Paciente", report.PatientName)
	writeOptionalLine(&builder, "Laboratorio", report.LabName)
	writeOptionalLine(&builder, "Solicitante", report.RequestingDoctor)
	writeDateLine(&builder, "Data do laudo", report.ReportDate)
	for _, result := range report.TestResults {
		builder.WriteByte('\n')
		writeLine(result.TestName)
		writeOptionalLine(&builder, "Material", result.Material)
		writeOptionalLine(&builder, "Metodo", result.Method)
		writeDateLine(&builder, "Coletado em", result.CollectedAt)
		writeDateLine(&builder, "Liberado em", result.ReleaseAt)
		for _, item := range result.Items {
			writeLine(formatLaboratoryItem(item))
		}
	}
	return strings.TrimSpace(builder.String())
}

func writeOptionalLine(builder *strings.Builder, label string, value *string) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return
	}
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(strings.TrimSpace(*value))
	builder.WriteByte('\n')
}

func writeDateLine(builder *strings.Builder, label string, value *time.Time) {
	if value == nil {
		return
	}
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(value.Format("02/01/2006"))
	builder.WriteByte('\n')
}

func formatLaboratoryItem(item laboratory.TestItemOutput) string {
	parts := []string{"-", strings.TrimSpace(item.ParameterName)}
	if item.ResultValue != nil && strings.TrimSpace(*item.ResultValue) != "" {
		parts = append(parts, strings.TrimSpace(*item.ResultValue))
	}
	if item.ResultUnit != nil && strings.TrimSpace(*item.ResultUnit) != "" {
		parts = append(parts, strings.TrimSpace(*item.ResultUnit))
	}
	if item.ReferenceText != nil && strings.TrimSpace(*item.ReferenceText) != "" {
		parts = append(parts, "(Referencia:", strings.TrimSpace(*item.ReferenceText)+")")
	}
	return strings.Join(parts, " ")
}
