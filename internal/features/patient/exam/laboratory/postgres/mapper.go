// internal/features/patient/exam/laboratory/postgres/mapper.go
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	labs "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	repohelpers "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres"
	labsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres/sqlc/generated/lab"
)

func mapLabReportRow(row labsqlc.GetLabReportByIDRow, panels []labs.LabPanel) labs.LabReport {
	return labs.LabReport{
		ID:                row.ID,
		ExamDocumentID:    pgUUIDToNullableUUID(row.ExamDocumentID),
		PatientID:         row.PatientID,
		PatientName:       repohelpers.FromPgTextToNullableString(row.PatientName),
		PatientDOB:        repohelpers.FromPgTimestamptzToNullableTimestamptz(row.PatientDob),
		LabName:           repohelpers.FromPgTextToNullableString(row.LabName),
		LabPhone:          repohelpers.FromPgTextToNullableString(row.LabPhone),
		InsuranceProvider: repohelpers.FromPgTextToNullableString(row.InsuranceProvider),
		RequestingDoctor:  repohelpers.FromPgTextToNullableString(row.RequestingDoctor),
		TechnicalManager:  repohelpers.FromPgTextToNullableString(row.TechnicalManager),
		ReportDate:        repohelpers.FromPgTimestamptzToNullableTimestamptz(row.ReportDate),
		TestResults:       panels,
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
		UploadedBy:        row.UploadedByUserID,
	}
}

func mapLabReportSummaryRow(row labsqlc.ListLabReportsByPatientIDRow) labs.LabReport {
	return labs.LabReport{
		ID:          row.ReportID,
		PatientID:   row.PatientID,
		PatientName: repohelpers.FromPgTextToNullableString(row.PatientName),
		LabName:     repohelpers.FromPgTextToNullableString(row.LabName),
		ReportDate:  repohelpers.FromPgTimestamptzToNullableTimestamptz(row.ReportDate),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		UploadedBy:  row.UploadedByUserID,
	}
}

func mapLabReportListRows(rows []labsqlc.ListLabReportsByPatientIDRow) []labs.LabReport {
	reports := make([]labs.LabReport, 0)
	reportIndexes := make(map[uuid.UUID]int)
	panelIndexes := make(map[uuid.UUID]int)

	for _, row := range rows {
		reportIndex, ok := reportIndexes[row.ReportID]
		if !ok {
			reportIndex = len(reports)
			reports = append(reports, mapLabReportSummaryRow(row))
			reportIndexes[row.ReportID] = reportIndex
		}

		if !row.PanelID.Valid {
			continue
		}

		panelID := uuid.UUID(row.PanelID.Bytes)
		panelIndex, ok := panelIndexes[panelID]
		if !ok {
			panelIndex = len(reports[reportIndex].TestResults)
			reports[reportIndex].TestResults = append(reports[reportIndex].TestResults, labs.LabPanel{
				ID:          panelID,
				LabReportID: reports[reportIndex].ID,
				TestName:    row.TestName.String,
				Material:    repohelpers.FromPgTextToNullableString(row.Material),
				Method:      repohelpers.FromPgTextToNullableString(row.Method),
				CollectedAt: repohelpers.FromPgTimestamptzToNullableTimestamptz(row.CollectedAt),
				ReleaseAt:   repohelpers.FromPgTimestamptzToNullableTimestamptz(row.ReleaseAt),
			})
			panelIndexes[panelID] = panelIndex
		}

		if !row.ObservationID.Valid {
			continue
		}
		observation := labs.Observation{
			ID:            uuid.UUID(row.ObservationID.Bytes),
			LabPanelID:    panelID,
			ParameterName: row.ParameterName.String,
			ResultValue:   repohelpers.FromPgTextToNullableString(row.ResultValue),
			ResultUnit:    repohelpers.FromPgTextToNullableString(row.ResultUnit),
			ReferenceText: repohelpers.FromPgTextToNullableString(row.ReferenceText),
		}
		reports[reportIndex].TestResults[panelIndex].Items = append(reports[reportIndex].TestResults[panelIndex].Items, observation)
	}

	return reports
}

func mapLabPanelRow(row labsqlc.LabPanel, observations []labs.Observation) labs.LabPanel {
	return labs.LabPanel{
		ID:          row.ID,
		LabReportID: row.LabReportID,
		TestName:    row.TestName,
		Material:    repohelpers.FromPgTextToNullableString(row.Material),
		Method:      repohelpers.FromPgTextToNullableString(row.Method),
		CollectedAt: repohelpers.FromPgTimestamptzToNullableTimestamptz(row.CollectedAt),
		ReleaseAt:   repohelpers.FromPgTimestamptzToNullableTimestamptz(row.ReleaseAt),
		Items:       observations,
	}
}

func mapObservationRow(row labsqlc.Observation) labs.Observation {
	return labs.Observation{
		ID:            row.ID,
		LabPanelID:    row.LabPanelID,
		ParameterName: row.ParameterName,
		ResultValue:   repohelpers.FromPgTextToNullableString(row.ResultValue),
		ResultUnit:    repohelpers.FromPgTextToNullableString(row.ResultUnit),
		ReferenceText: repohelpers.FromPgTextToNullableString(row.ReferenceText),
	}
}

func mapObservationTimelineRow(row labsqlc.ListObservationTimelineByPatientAndParameterRow) labs.ObservationTimeline {
	return labs.ObservationTimeline{
		ReportID:      row.ReportID,
		LabPanelID:    row.LabPanelID,
		ObservationID: row.ObservationID,
		ReportDate:    repohelpers.FromPgTimestamptzToNullableTimestamptz(row.ReportDate),
		TestName:      row.TestName,
		ParameterName: row.ParameterName,
		ResultValue:   repohelpers.FromPgTextToNullableString(row.ResultValue),
		ResultUnit:    repohelpers.FromPgTextToNullableString(row.ResultUnit),
		ReferenceText: repohelpers.FromPgTextToNullableString(row.ReferenceText),
	}
}

func nullableStringToPg(value *string) pgtype.Text {
	return repohelpers.FromNullableStringToPgText(value)
}

func nullableTimeToPg(value *time.Time) pgtype.Timestamptz {
	return repohelpers.FromNullableTimestamptzToPgTimestamptz(value)
}

func labReportDocumentIDToPg(value *uuid.UUID) pgtype.UUID {
	return repohelpers.FromNullableUUIDToPgUUID(value)
}

func pgUUIDToNullableUUID(value pgtype.UUID) *uuid.UUID {
	return repohelpers.FromPgUUIDToNullableUUID(value)
}
