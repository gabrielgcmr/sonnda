// internal/features/patient/exam/laboratory/postgres/repository.go
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	repohelpers "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres"

	labrepository "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	labs "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres"
	labsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres/sqlc/generated/lab"
	"github.com/gabrielgcmr/sonnda/internal/kernel/persistence"

	"github.com/google/uuid"
)

type Repository struct {
	client  *postgress.Client
	queries *labsqlc.Queries
}

var _ labrepository.Repository = (*Repository)(nil)

func NewRepository(client *postgress.Client) *Repository {
	return &Repository{
		client:  client,
		queries: labsqlc.New(client.Pool()),
	}
}

// Create implements [repository.LabsRepository].
func (l *Repository) Create(ctx context.Context, report *labs.LabReport) error {
	if report == nil {
		return persistence.ErrPersistenceFailure
	}

	tx, err := l.client.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// O rollback precisa funcionar mesmo se a requisicao foi cancelada.
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanupCtx)
	}()
	if err := l.CreateInTx(ctx, tx, report); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// CreateInTx lets a cross-feature use case commit the clinical report with its document.
func (l *Repository) CreateInTx(ctx context.Context, tx pgx.Tx, report *labs.LabReport) error {
	if report == nil {
		return persistence.ErrPersistenceFailure
	}
	queries := l.queries.WithTx(tx)

	// Report, panels, and observations are created atomically; document metadata belongs to documentprocessing.
	_, err := queries.CreateLabReport(ctx, labsqlc.CreateLabReportParams{
		ID:                report.ID,
		PatientID:         report.PatientID,
		ExamDocumentID:    nullableUUIDToPg(report.ExamDocumentID),
		PatientName:       repohelpers.FromNullableStringToPgText(report.PatientName),
		PatientDob:        repohelpers.FromNullableTimestamptzToPgTimestamptz(report.PatientDOB),
		LabName:           repohelpers.FromNullableStringToPgText(report.LabName),
		LabPhone:          repohelpers.FromNullableStringToPgText(report.LabPhone),
		InsuranceProvider: repohelpers.FromNullableStringToPgText(report.InsuranceProvider),
		RequestingDoctor:  repohelpers.FromNullableStringToPgText(report.RequestingDoctor),
		TechnicalManager:  repohelpers.FromNullableStringToPgText(report.TechnicalManager),
		ReportDate:        repohelpers.FromNullableTimestamptzToPgTimestamptz(report.ReportDate),
		UploadedByUserID:  report.UploadedBy,
	})
	if err != nil {
		return err
	}

	// Create panels and their observations.
	for _, tr := range report.TestResults {
		_, err := queries.CreateLabPanel(ctx, labsqlc.CreateLabPanelParams{
			ID:          tr.ID,
			LabReportID: report.ID,
			TestName:    tr.TestName,
			Material:    repohelpers.FromNullableStringToPgText(tr.Material),
			Method:      repohelpers.FromNullableStringToPgText(tr.Method),
			CollectedAt: repohelpers.FromNullableTimestamptzToPgTimestamptz(tr.CollectedAt),
			ReleaseAt:   repohelpers.FromNullableTimestamptzToPgTimestamptz(tr.ReleaseAt),
		})
		if err != nil {
			return err
		}

		for _, item := range tr.Items {
			_, err := queries.CreateObservation(ctx, labsqlc.CreateObservationParams{
				ID:            item.ID,
				LabPanelID:    tr.ID,
				ParameterName: item.ParameterName,
				ResultValue:   repohelpers.FromNullableStringToPgText(item.ResultValue),
				ResultUnit:    repohelpers.FromNullableStringToPgText(item.ResultUnit),
				ReferenceText: repohelpers.FromNullableStringToPgText(item.ReferenceText),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete implements [repository.LabsRepository].
func (l *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	// Foreign-key cascades remove panels and observations with the report.
	_, err := l.queries.DeleteLabReport(ctx, id)
	return err
}

// FindByID implements [repository.LabsRepository].
func (l *Repository) FindByID(ctx context.Context, reportID uuid.UUID) (*labs.LabReport, error) {
	reportRow, err := l.queries.GetLabReportByID(ctx, reportID)
	if err != nil {
		if IsNoRows(err) {
			return nil, nil
		}
		return nil, err
	}

	// Fetch panels.
	resultsRows, err := l.queries.ListLabPanelsByReportID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	var testResults []labs.LabPanel
	for _, resultRow := range resultsRows {
		observationsRows, err := l.queries.ListObservationsByPanelID(ctx, resultRow.ID)
		if err != nil {
			return nil, err
		}

		var observations []labs.Observation
		for _, observationRow := range observationsRows {
			observations = append(observations, labs.Observation{
				ID:            observationRow.ID,
				LabPanelID:    observationRow.LabPanelID,
				ParameterName: observationRow.ParameterName,
				ResultValue:   repohelpers.FromPgTextToNullableString(observationRow.ResultValue),
				ResultUnit:    repohelpers.FromPgTextToNullableString(observationRow.ResultUnit),
				ReferenceText: repohelpers.FromPgTextToNullableString(observationRow.ReferenceText),
			})
		}

		testResults = append(testResults, labs.LabPanel{
			ID:          resultRow.ID,
			LabReportID: resultRow.LabReportID,
			TestName:    resultRow.TestName,
			Material:    repohelpers.FromPgTextToNullableString(resultRow.Material),
			Method:      repohelpers.FromPgTextToNullableString(resultRow.Method),
			CollectedAt: repohelpers.FromPgTimestamptzToNullableTimestamptz(resultRow.CollectedAt),
			ReleaseAt:   repohelpers.FromPgTimestamptzToNullableTimestamptz(resultRow.ReleaseAt),
			Items:       observations,
		})
	}

	return &labs.LabReport{
		ID:                reportRow.ID,
		ExamDocumentID:    nullableReportDocumentID(reportRow.ExamDocumentID),
		PatientID:         reportRow.PatientID,
		PatientName:       repohelpers.FromPgTextToNullableString(reportRow.PatientName),
		PatientDOB:        repohelpers.FromPgTimestamptzToNullableTimestamptz(reportRow.PatientDob),
		LabName:           repohelpers.FromPgTextToNullableString(reportRow.LabName),
		LabPhone:          repohelpers.FromPgTextToNullableString(reportRow.LabPhone),
		InsuranceProvider: repohelpers.FromPgTextToNullableString(reportRow.InsuranceProvider),
		RequestingDoctor:  repohelpers.FromPgTextToNullableString(reportRow.RequestingDoctor),
		TechnicalManager:  repohelpers.FromPgTextToNullableString(reportRow.TechnicalManager),
		ReportDate:        repohelpers.FromPgTimestamptzToNullableTimestamptz(reportRow.ReportDate),
		TestResults:       testResults,
		CreatedAt:         reportRow.CreatedAt.Time,
		UpdatedAt:         reportRow.UpdatedAt.Time,
		UploadedBy:        reportRow.UploadedByUserID,
	}, nil
}

// ListObservationTimelineByPatientAndParameter implements [repository.Repository].
func (l *Repository) ListObservationTimelineByPatientAndParameter(ctx context.Context, patientID uuid.UUID, parameterName string, limit int, offset int) ([]labs.ObservationTimeline, error) {
	rows, err := l.queries.ListObservationTimelineByPatientAndParameter(ctx, labsqlc.ListObservationTimelineByPatientAndParameterParams{
		PatientID:     patientID,
		ParameterName: parameterName,
		Limit:         int32(limit),
		Offset:        int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var observations []labs.ObservationTimeline
	for _, row := range rows {
		observations = append(observations, labs.ObservationTimeline{
			ReportID:      row.ReportID,
			LabPanelID:    row.LabPanelID,
			ObservationID: row.ObservationID,
			ReportDate:    repohelpers.FromPgTimestamptzToNullableTimestamptz(row.ReportDate),
			TestName:      row.TestName,
			ParameterName: row.ParameterName,
			ResultValue:   repohelpers.FromPgTextToNullableString(row.ResultValue),
			ResultUnit:    repohelpers.FromPgTextToNullableString(row.ResultUnit),
			ReferenceText: repohelpers.FromPgTextToNullableString(row.ReferenceText),
		})
	}

	return observations, nil
}

// ListLabs implements [repository.LabsRepository].
func (l *Repository) ListLabs(ctx context.Context, patientID uuid.UUID, limit int, offset int) ([]labs.LabReport, error) {
	rows, err := l.queries.ListLabReportsByPatientID(ctx, labsqlc.ListLabReportsByPatientIDParams{
		PatientID: patientID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var reports []labs.LabReport
	for _, row := range rows {
		reports = append(reports, labs.LabReport{
			ID:          row.ID,
			PatientID:   row.PatientID,
			PatientName: repohelpers.FromPgTextToNullableString(row.PatientName),
			LabName:     repohelpers.FromPgTextToNullableString(row.LabName),
			ReportDate:  repohelpers.FromPgTimestamptzToNullableTimestamptz(row.ReportDate),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			UploadedBy:  row.UploadedByUserID,
		})
	}

	return reports, nil
}

func nullableReportDocumentID(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	id := uuid.UUID(value.Bytes)
	return &id
}

func nullableUUIDToPg(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *value, Valid: true}
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
