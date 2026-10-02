// internal/features/patient/exam/laboratory/postgres/read.go
package postgres

import (
	"context"

	labs "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	labsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres/sqlc/generated/lab"
	"github.com/google/uuid"
)

// FindByID implements [laboratory.Repository].
func (r *Repository) FindByID(ctx context.Context, reportID uuid.UUID) (*labs.LabReport, error) {
	reportRow, err := r.queries.GetLabReportByID(ctx, reportID)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, persistenceError("get laboratory report", err)
	}

	panelRows, err := r.queries.ListLabPanelsByReportID(ctx, reportID)
	if err != nil {
		return nil, persistenceError("list laboratory panels", err)
	}

	panels := make([]labs.LabPanel, 0, len(panelRows))
	for _, panelRow := range panelRows {
		observationRows, err := r.queries.ListObservationsByPanelID(ctx, panelRow.ID)
		if err != nil {
			return nil, persistenceError("list laboratory observations", err)
		}

		observations := make([]labs.Observation, 0, len(observationRows))
		for _, observationRow := range observationRows {
			observations = append(observations, mapObservationRow(observationRow))
		}
		panels = append(panels, mapLabPanelRow(panelRow, observations))
	}

	report := mapLabReportRow(reportRow, panels)
	return &report, nil
}

// ListObservationTimelineByPatientAndParameter implements [laboratory.Repository].
func (r *Repository) ListObservationTimelineByPatientAndParameter(ctx context.Context, patientID uuid.UUID, parameterName string, limit int, offset int) ([]labs.ObservationTimeline, error) {
	rows, err := r.queries.ListObservationTimelineByPatientAndParameter(ctx, labsqlc.ListObservationTimelineByPatientAndParameterParams{
		PatientID:     patientID,
		ParameterName: parameterName,
		Limit:         int32(limit),
		Offset:        int32(offset),
	})
	if err != nil {
		return nil, persistenceError("list observation timeline", err)
	}

	observations := make([]labs.ObservationTimeline, 0, len(rows))
	for _, row := range rows {
		observations = append(observations, mapObservationTimelineRow(row))
	}
	return observations, nil
}

// ListLabs implements [laboratory.Repository].
func (r *Repository) ListLabs(ctx context.Context, patientID uuid.UUID, limit int, offset int) ([]labs.LabReport, error) {
	rows, err := r.queries.ListLabReportsByPatientID(ctx, labsqlc.ListLabReportsByPatientIDParams{
		PatientID: patientID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, persistenceError("list laboratory reports", err)
	}

	return mapLabReportListRows(rows), nil
}
