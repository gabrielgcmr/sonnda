// internal/features/patient/exam/laboratory/postgres/mapper_test.go
package postgres

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	labsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres/sqlc/generated/lab"
)

func TestMapLabReportListRowsGroupsPanelsAndObservations(t *testing.T) {
	reportID := uuid.New()
	patientID := uuid.New()
	panelID := uuid.New()
	firstObservationID := uuid.New()
	secondObservationID := uuid.New()

	rows := []labsqlc.ListLabReportsByPatientIDRow{
		{
			ReportID:      reportID,
			PatientID:     patientID,
			PanelID:       pgtype.UUID{Bytes: panelID, Valid: true},
			TestName:      pgtype.Text{String: "Hemograma", Valid: true},
			ObservationID: pgtype.UUID{Bytes: firstObservationID, Valid: true},
			ParameterName: pgtype.Text{String: "Hemoglobina", Valid: true},
		},
		{
			ReportID:      reportID,
			PatientID:     patientID,
			PanelID:       pgtype.UUID{Bytes: panelID, Valid: true},
			TestName:      pgtype.Text{String: "Hemograma", Valid: true},
			ObservationID: pgtype.UUID{Bytes: secondObservationID, Valid: true},
			ParameterName: pgtype.Text{String: "Leucócitos", Valid: true},
			ResultValue:   pgtype.Text{String: "7.2", Valid: true},
			ResultUnit:    pgtype.Text{String: "mil/mm³", Valid: true},
			ReferenceText: pgtype.Text{String: "4-11", Valid: true},
		},
	}

	reports := mapLabReportListRows(rows)
	if len(reports) != 1 {
		t.Fatalf("expected one report, got %d", len(reports))
	}
	if len(reports[0].TestResults) != 1 {
		t.Fatalf("expected one panel, got %d", len(reports[0].TestResults))
	}
	if len(reports[0].TestResults[0].Items) != 2 {
		t.Fatalf("expected two observations, got %d", len(reports[0].TestResults[0].Items))
	}
}
