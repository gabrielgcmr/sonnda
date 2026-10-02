// internal/features/patient/exam/laboratory/postgres/write.go
package postgres

import (
	"context"
	"time"

	labs "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	labsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/database/postgres/sqlc/generated/lab"
	"github.com/gabrielgcmr/sonnda/internal/kernel/persistence"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Create implements [laboratory.Repository].
func (r *Repository) Create(ctx context.Context, report *labs.LabReport) error {
	if report == nil {
		return persistence.ErrPersistenceFailure
	}

	tx, err := r.client.BeginTx(ctx)
	if err != nil {
		return persistenceError("begin laboratory transaction", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanupCtx)
	}()

	if err := r.CreateInTx(ctx, tx, report); err != nil {
		return err
	}
	return persistenceError("commit laboratory transaction", tx.Commit(ctx))
}

// CreateInTx writes a report aggregate using the caller's transaction.
func (r *Repository) CreateInTx(ctx context.Context, tx pgx.Tx, report *labs.LabReport) error {
	if report == nil {
		return persistence.ErrPersistenceFailure
	}

	queries := r.queries.WithTx(tx)
	if _, err := queries.CreateLabReport(ctx, labsqlc.CreateLabReportParams{
		ID:                report.ID,
		PatientID:         report.PatientID,
		ExamDocumentID:    labReportDocumentIDToPg(report.ExamDocumentID),
		PatientName:       nullableStringToPg(report.PatientName),
		PatientDob:        nullableTimeToPg(report.PatientDOB),
		LabName:           nullableStringToPg(report.LabName),
		LabPhone:          nullableStringToPg(report.LabPhone),
		InsuranceProvider: nullableStringToPg(report.InsuranceProvider),
		RequestingDoctor:  nullableStringToPg(report.RequestingDoctor),
		TechnicalManager:  nullableStringToPg(report.TechnicalManager),
		ReportDate:        nullableTimeToPg(report.ReportDate),
		UploadedByUserID:  report.UploadedBy,
	}); err != nil {
		return persistenceError("create laboratory report", err)
	}

	for _, panel := range report.TestResults {
		if _, err := queries.CreateLabPanel(ctx, labsqlc.CreateLabPanelParams{
			ID:          panel.ID,
			LabReportID: report.ID,
			TestName:    panel.TestName,
			Material:    nullableStringToPg(panel.Material),
			Method:      nullableStringToPg(panel.Method),
			CollectedAt: nullableTimeToPg(panel.CollectedAt),
			ReleaseAt:   nullableTimeToPg(panel.ReleaseAt),
		}); err != nil {
			return persistenceError("create laboratory panel", err)
		}

		for _, observation := range panel.Items {
			if _, err := queries.CreateObservation(ctx, labsqlc.CreateObservationParams{
				ID:            observation.ID,
				LabPanelID:    panel.ID,
				ParameterName: observation.ParameterName,
				ResultValue:   nullableStringToPg(observation.ResultValue),
				ResultUnit:    nullableStringToPg(observation.ResultUnit),
				ReferenceText: nullableStringToPg(observation.ReferenceText),
			}); err != nil {
				return persistenceError("create laboratory observation", err)
			}
		}
	}

	return nil
}

// Delete implements [laboratory.Repository].
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.DeleteLabReport(ctx, id)
	return persistenceError("delete laboratory report", err)
}
