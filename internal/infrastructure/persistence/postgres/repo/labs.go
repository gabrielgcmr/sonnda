// internal/infrastructure/persistence/postgres/repo/labs.go
// internal/adapters/outbound/data/postgres/labs.go
package repo

import (
	"context"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	labsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/lab"

	"github.com/google/uuid"
)

type LabsRepository struct {
	client  *postgress.Client
	queries *labsqlc.Queries
}

var _ repository.Labs = (*LabsRepository)(nil)

func NewLabsRepository(client *postgress.Client) repository.Labs {
	return &LabsRepository{
		client:  client,
		queries: labsqlc.New(client.Pool()),
	}
}

// Create implements [repository.LabsRepository].
func (l *LabsRepository) Create(ctx context.Context, report *labs.LabReport) error {
	if report == nil {
		return ErrRepositoryFailure
	}

	// Create the lab report
	reportRow, err := l.queries.CreateLabReport(ctx, labsqlc.CreateLabReportParams{
		ID:                report.ID,
		PatientID:         report.PatientID,
		PatientName:       FromNullableStringToPgText(report.PatientName),
		PatientDob:        FromNullableTimestamptzToPgTimestamptz(report.PatientDOB),
		LabName:           FromNullableStringToPgText(report.LabName),
		LabPhone:          FromNullableStringToPgText(report.LabPhone),
		InsuranceProvider: FromNullableStringToPgText(report.InsuranceProvider),
		RequestingDoctor:  FromNullableStringToPgText(report.RequestingDoctor),
		TechnicalManager:  FromNullableStringToPgText(report.TechnicalManager),
		ReportDate:        FromNullableTimestamptzToPgTimestamptz(report.ReportDate),
		RawText:           FromNullableStringToPgText(report.RawText),
		UploadedByUserID:  report.UploadedBy,
		Fingerprint:       FromNullableStringToPgText(report.Fingerprint),
	})
	if err != nil {
		return err
	}

	// Create test results and their items
	for _, tr := range report.TestResults {
		_, err := l.queries.CreateLabResult(ctx, labsqlc.CreateLabResultParams{
			ID:          tr.ID,
			LabReportID: reportRow.ID,
			TestName:    tr.TestName,
			Material:    FromNullableStringToPgText(tr.Material),
			Method:      FromNullableStringToPgText(tr.Method),
			CollectedAt: FromNullableTimestamptzToPgTimestamptz(tr.CollectedAt),
			ReleaseAt:   FromNullableTimestamptzToPgTimestamptz(tr.ReleaseAt),
		})
		if err != nil {
			return err
		}

		for _, item := range tr.Items {
			_, err := l.queries.CreateLabResultItem(ctx, labsqlc.CreateLabResultItemParams{
				ID:            item.ID,
				LabResultID:   item.LabResultID,
				ParameterName: item.ParameterName,
				ResultValue:   FromNullableStringToPgText(item.ResultValue),
				ResultUnit:    FromNullableStringToPgText(item.ResultUnit),
				ReferenceText: FromNullableStringToPgText(item.ReferenceText),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
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

// ExistsBySignature implements [repository.LabsRepository].
func (l *LabsRepository) ExistsBySignature(ctx context.Context, patientID uuid.UUID, fingerprint string) (bool, error) {
	exists, err := l.queries.ExistsLabReportByPatientAndFingerprint(ctx, labsqlc.ExistsLabReportByPatientAndFingerprintParams{
		PatientID:   patientID,
		Fingerprint: FromRequiredStringToPgText(fingerprint),
	})
	return exists, err
}

// FindByID implements [repository.LabsRepository].
func (l *LabsRepository) FindByID(ctx context.Context, reportID uuid.UUID) (*labs.LabReport, error) {
	reportRow, err := l.queries.GetLabReportByID(ctx, reportID)
	if err != nil {
		if IsPgNotFound(err) {
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
				ResultValue:   FromPgTextToNullableString(itemRow.ResultValue),
				ResultUnit:    FromPgTextToNullableString(itemRow.ResultUnit),
				ReferenceText: FromPgTextToNullableString(itemRow.ReferenceText),
			})
		}

		testResults = append(testResults, labs.LabResult{
			ID:          resultRow.ID,
			LabReportID: resultRow.LabReportID,
			TestName:    resultRow.TestName,
			Material:    FromPgTextToNullableString(resultRow.Material),
			Method:      FromPgTextToNullableString(resultRow.Method),
			CollectedAt: FromPgTimestamptzToNullableTimestamptz(resultRow.CollectedAt),
			ReleaseAt:   FromPgTimestamptzToNullableTimestamptz(resultRow.ReleaseAt),
			Items:       items,
		})
	}

	return &labs.LabReport{
		ID:                reportRow.ID,
		PatientID:         reportRow.PatientID,
		PatientName:       FromPgTextToNullableString(reportRow.PatientName),
		PatientDOB:        FromPgTimestamptzToNullableTimestamptz(reportRow.PatientDob),
		LabName:           FromPgTextToNullableString(reportRow.LabName),
		LabPhone:          FromPgTextToNullableString(reportRow.LabPhone),
		InsuranceProvider: FromPgTextToNullableString(reportRow.InsuranceProvider),
		RequestingDoctor:  FromPgTextToNullableString(reportRow.RequestingDoctor),
		TechnicalManager:  FromPgTextToNullableString(reportRow.TechnicalManager),
		ReportDate:        FromPgTimestamptzToNullableTimestamptz(reportRow.ReportDate),
		Fingerprint:       FromPgTextToNullableString(reportRow.Fingerprint),
		RawText:           FromPgTextToNullableString(reportRow.RawText),
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
			ReportDate:    FromPgTimestamptzToNullableTimestamptz(row.ReportDate),
			TestName:      row.TestName,
			ParameterName: row.ParameterName,
			ResultValue:   FromPgTextToNullableString(row.ResultValue),
			ResultUnit:    FromPgTextToNullableString(row.ResultUnit),
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
			PatientName: FromPgTextToNullableString(row.PatientName),
			LabName:     FromPgTextToNullableString(row.LabName),
			ReportDate:  FromPgTimestamptzToNullableTimestamptz(row.ReportDate),
			Fingerprint: FromPgTextToNullableString(row.Fingerprint),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			UploadedBy:  row.UploadedByUserID,
		})
	}

	return reports, nil
}
