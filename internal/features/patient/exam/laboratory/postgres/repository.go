// internal/features/patient/exam/laboratory/postgres/repository.go
package postgres

import (
	"context"
	"time"

	repohelpers "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/repo"

	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	labrepository "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	labs "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	labsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/lab"

	"github.com/google/uuid"
)

type LabsRepository struct {
	client  *postgress.Client
	queries *labsqlc.Queries
}

var _ labrepository.Repository = (*LabsRepository)(nil)

func NewLabsRepository(client *postgress.Client) *LabsRepository {
	return &LabsRepository{
		client:  client,
		queries: labsqlc.New(client.Pool()),
	}
}

// Create implements [repository.LabsRepository].
func (l *LabsRepository) Create(ctx context.Context, report *labs.LabReport) error {
	if report == nil {
		return repository.ErrRepositoryFailure
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
	queries := l.queries.WithTx(tx)

	// Laudo, resultados e itens sao gravados juntos; metadata documental pertence a documentprocessing.
	_, err = tx.Exec(ctx, `
		INSERT INTO lab_reports (
			id, patient_id, patient_name, patient_dob, lab_name, lab_phone,
			insurance_provider, requesting_doctor, technical_manager, report_date, uploaded_by_user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		report.ID,
		report.PatientID,
		repohelpers.FromNullableStringToPgText(report.PatientName),
		repohelpers.FromNullableTimestamptzToPgTimestamptz(report.PatientDOB),
		repohelpers.FromNullableStringToPgText(report.LabName),
		repohelpers.FromNullableStringToPgText(report.LabPhone),
		repohelpers.FromNullableStringToPgText(report.InsuranceProvider),
		repohelpers.FromNullableStringToPgText(report.RequestingDoctor),
		repohelpers.FromNullableStringToPgText(report.TechnicalManager),
		repohelpers.FromNullableTimestamptzToPgTimestamptz(report.ReportDate),
		report.UploadedBy,
	)
	if err != nil {
		return err
	}

	// Create test results and their items
	for _, tr := range report.TestResults {
		_, err := queries.CreateLabResult(ctx, labsqlc.CreateLabResultParams{
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
			_, err := queries.CreateLabResultItem(ctx, labsqlc.CreateLabResultItemParams{
				ID:            item.ID,
				LabResultID:   tr.ID,
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

	return tx.Commit(ctx)
}

// Delete implements [repository.LabsRepository].
func (l *LabsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete items first
	_, err := l.queries.DeleteLabResultItemsByReportID(ctx, id)
	if err != nil {
		return err
	}

	// Then delete results
	_, err = l.queries.DeleteLabResultsByReportID(ctx, id)
	if err != nil {
		return err
	}

	// Finally delete the report
	_, err = l.queries.DeleteLabReport(ctx, id)
	return err
}

// FindByID implements [repository.LabsRepository].
func (l *LabsRepository) FindByID(ctx context.Context, reportID uuid.UUID) (*labs.LabReport, error) {
	reportRow, err := l.queries.GetLabReportByID(ctx, reportID)
	if err != nil {
		if repohelpers.IsPgNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// Fetch test results
	resultsRows, err := l.queries.ListLabResultsByReportID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	var testResults []labs.LabResult
	for _, resultRow := range resultsRows {
		itemsRows, err := l.queries.ListLabResultItemsByResultID(ctx, resultRow.ID)
		if err != nil {
			return nil, err
		}

		var items []labs.LabResultItem
		for _, itemRow := range itemsRows {
			items = append(items, labs.LabResultItem{
				ID:            itemRow.ID,
				LabResultID:   itemRow.LabResultID,
				ParameterName: itemRow.ParameterName,
				ResultValue:   repohelpers.FromPgTextToNullableString(itemRow.ResultValue),
				ResultUnit:    repohelpers.FromPgTextToNullableString(itemRow.ResultUnit),
				ReferenceText: repohelpers.FromPgTextToNullableString(itemRow.ReferenceText),
			})
		}

		testResults = append(testResults, labs.LabResult{
			ID:          resultRow.ID,
			LabReportID: resultRow.LabReportID,
			TestName:    resultRow.TestName,
			Material:    repohelpers.FromPgTextToNullableString(resultRow.Material),
			Method:      repohelpers.FromPgTextToNullableString(resultRow.Method),
			CollectedAt: repohelpers.FromPgTimestamptzToNullableTimestamptz(resultRow.CollectedAt),
			ReleaseAt:   repohelpers.FromPgTimestamptzToNullableTimestamptz(resultRow.ReleaseAt),
			Items:       items,
		})
	}

	return &labs.LabReport{
		ID:                reportRow.ID,
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

// ListItemsByPatientAndParameter implements [repository.LabsRepository].
func (l *LabsRepository) ListItemsByPatientAndParameter(ctx context.Context, patientID uuid.UUID, parameterName string, limit int, offset int) ([]labs.LabResultItemTimeline, error) {
	rows, err := l.queries.ListLabItemTimelineByPatientAndParameter(ctx, labsqlc.ListLabItemTimelineByPatientAndParameterParams{
		PatientID:     patientID,
		ParameterName: parameterName,
		Limit:         int32(limit),
		Offset:        int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var items []labs.LabResultItemTimeline
	for _, row := range rows {
		items = append(items, labs.LabResultItemTimeline{
			ReportID:      row.ReportID,
			LabResultID:   row.LabResultID,
			ItemID:        row.ItemID,
			ReportDate:    repohelpers.FromPgTimestamptzToNullableTimestamptz(row.ReportDate),
			TestName:      row.TestName,
			ParameterName: row.ParameterName,
			ResultValue:   repohelpers.FromPgTextToNullableString(row.ResultValue),
			ResultUnit:    repohelpers.FromPgTextToNullableString(row.ResultUnit),
		})
	}

	return items, nil
}

// ListLabs implements [repository.LabsRepository].
func (l *LabsRepository) ListLabs(ctx context.Context, patientID uuid.UUID, limit int, offset int) ([]labs.LabReport, error) {
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
