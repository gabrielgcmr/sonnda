// internal/infrastructure/persistence/postgres/repo/labs.go
// internal/adapters/outbound/data/postgres/labs.go
package repo

import (
	"context"
	"errors"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	labsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/lab"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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

	if report.ExamDocumentID != nil {
		valid, err := queries.LabDocumentBelongsToPatient(ctx, labsqlc.LabDocumentBelongsToPatientParams{
			ID: *report.ExamDocumentID, PatientID: report.PatientID,
		})
		if err != nil {
			return err
		}
		if !valid {
			return labs.ErrDocumentLinkConflict
		}
	}

	// Laudo, resultados e itens sao gravados juntos.
	reportRow, err := queries.CreateLabReport(ctx, labsqlc.CreateLabReportParams{
		ID:                report.ID,
		PatientID:         report.PatientID,
		ExamDocumentID:    FromNullableUUIDToPgUUID(report.ExamDocumentID),
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
		return mapLabWriteError(err)
	}

	// Create test results and their items
	for _, tr := range report.TestResults {
		_, err := queries.CreateLabResult(ctx, labsqlc.CreateLabResultParams{
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
			_, err := queries.CreateLabResultItem(ctx, labsqlc.CreateLabResultItemParams{
				ID:            item.ID,
				LabResultID:   tr.ID,
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

	return tx.Commit(ctx)
}

func (l *LabsRepository) AttachDocument(ctx context.Context, reportID, patientID, documentID uuid.UUID) error {
	rows, err := l.queries.AttachLabReportDocument(ctx, labsqlc.AttachLabReportDocumentParams{
		ReportID: reportID, PatientID: patientID, DocumentID: documentID,
	})
	if err != nil {
		return mapLabWriteError(err)
	}
	if rows == 0 {
		return labs.ErrDocumentLinkConflict
	}
	return nil
}

func mapLabWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "idx_lab_reports_fingerprint", "idx_lab_reports_exam_document":
			return errors.Join(labs.ErrLabReportAlreadyExists, err)
		}
	}
	return err
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

// FindBySignature implements [repository.LabsRepository].
func (l *LabsRepository) FindBySignature(ctx context.Context, patientID uuid.UUID, fingerprint string) (*labs.LabReport, error) {
	reportRow, err := l.queries.GetLabReportByPatientAndFingerprint(ctx, labsqlc.GetLabReportByPatientAndFingerprintParams{
		PatientID:   patientID,
		Fingerprint: FromRequiredStringToPgText(fingerprint),
	})
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	resultsRows, err := l.queries.ListLabResultsByReportID(ctx, reportRow.ID)
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
		ExamDocumentID:    FromPgUUIDToNullableUUID(reportRow.ExamDocumentID),
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
		ExamDocumentID:    FromPgUUIDToNullableUUID(reportRow.ExamDocumentID),
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
			ID:             row.ID,
			PatientID:      row.PatientID,
			ExamDocumentID: FromPgUUIDToNullableUUID(row.ExamDocumentID),
			PatientName:    FromPgTextToNullableString(row.PatientName),
			LabName:        FromPgTextToNullableString(row.LabName),
			ReportDate:     FromPgTimestamptzToNullableTimestamptz(row.ReportDate),
			Fingerprint:    FromPgTextToNullableString(row.Fingerprint),
			CreatedAt:      row.CreatedAt.Time,
			UpdatedAt:      row.UpdatedAt.Time,
			UploadedBy:     row.UploadedByUserID,
		})
	}

	return reports, nil
}
