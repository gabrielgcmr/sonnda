// internal/features/documentprocessing/postgres/repository.go
package postgres

import (
	"context"
	"errors"

	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	processing "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	processingdomain "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	laboratory "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory"
	labdomain "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	pginfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository struct {
	client     *pginfra.Client
	laboratory laboratory.Repository
}

var _ processing.LaboratoryReportRepository = (*Repository)(nil)
var _ laboratory.Repository = (*Repository)(nil)

func NewRepository(client *pginfra.Client, laboratoryRepository laboratory.Repository) *Repository {
	return &Repository{client: client, laboratory: laboratoryRepository}
}

func (r *Repository) Create(ctx context.Context, report *labdomain.LabReport) error {
	return r.laboratory.Create(ctx, report)
}

func (r *Repository) CreateFromDocument(
	ctx context.Context,
	report *labdomain.LabReport,
	metadata processingdomain.LaboratoryReportMetadata,
) error {
	if report == nil {
		return repository.ErrRepositoryFailure
	}
	if metadata.ExamDocumentID != nil {
		var belongs bool
		if err := r.client.Pool().QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM exam_documents WHERE id = $1 AND patient_id = $2)`,
			*metadata.ExamDocumentID, report.PatientID,
		).Scan(&belongs); err != nil {
			return err
		}
		if !belongs {
			return labdomain.ErrDocumentLinkConflict
		}
	}

	if err := r.laboratory.Create(ctx, report); err != nil {
		return err
	}
	if err := r.persistMetadata(ctx, report, metadata); err != nil {
		cleanupErr := r.laboratory.Delete(ctx, report.ID)
		return mapProcessingWriteError(errors.Join(err, cleanupErr))
	}
	report.ExamDocumentID = metadata.ExamDocumentID
	return nil
}

func (r *Repository) FindByFingerprint(ctx context.Context, patientID uuid.UUID, fingerprint string) (*labdomain.LabReport, error) {
	var reportID uuid.UUID
	err := r.client.Pool().QueryRow(ctx,
		`SELECT id FROM lab_reports WHERE patient_id = $1 AND fingerprint = $2`,
		patientID, fingerprint,
	).Scan(&reportID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, reportID)
}

func (r *Repository) FindByID(ctx context.Context, reportID uuid.UUID) (*labdomain.LabReport, error) {
	report, err := r.laboratory.FindByID(ctx, reportID)
	if err != nil || report == nil {
		return report, err
	}
	if err := r.loadDocumentLink(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (r *Repository) AttachDocument(ctx context.Context, reportID, patientID, documentID uuid.UUID) error {
	result, err := r.client.Pool().Exec(ctx, `
		UPDATE lab_reports AS report
		SET exam_document_id = document.id, updated_at = now()
		FROM exam_documents AS document
		WHERE report.id = $1
		  AND report.patient_id = $2
		  AND document.id = $3
		  AND document.patient_id = report.patient_id
		  AND (report.exam_document_id IS NULL OR report.exam_document_id = document.id)`,
		reportID, patientID, documentID,
	)
	if err != nil {
		return mapProcessingWriteError(err)
	}
	if result.RowsAffected() == 0 {
		return labdomain.ErrDocumentLinkConflict
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.laboratory.Delete(ctx, id)
}

func (r *Repository) ListLabs(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]labdomain.LabReport, error) {
	reports, err := r.laboratory.ListLabs(ctx, patientID, limit, offset)
	if err != nil {
		return nil, err
	}
	for index := range reports {
		if err := r.loadDocumentLink(ctx, &reports[index]); err != nil {
			return nil, err
		}
	}
	return reports, nil
}

func (r *Repository) ListItemsByPatientAndParameter(ctx context.Context, patientID uuid.UUID, parameterName string, limit, offset int) ([]labdomain.LabResultItemTimeline, error) {
	return r.laboratory.ListItemsByPatientAndParameter(ctx, patientID, parameterName, limit, offset)
}

func (r *Repository) persistMetadata(ctx context.Context, report *labdomain.LabReport, metadata processingdomain.LaboratoryReportMetadata) error {
	result, err := r.client.Pool().Exec(ctx, `
		UPDATE lab_reports
		SET raw_text = $3, fingerprint = $4, exam_document_id = $5, updated_at = now()
		WHERE id = $1 AND patient_id = $2`,
		report.ID, report.PatientID, metadata.RawText, metadata.Fingerprint, metadata.ExamDocumentID,
	)
	if err != nil {
		return mapProcessingWriteError(err)
	}
	if result.RowsAffected() == 0 {
		return repository.ErrRepositoryFailure
	}
	return nil
}

func (r *Repository) loadDocumentLink(ctx context.Context, report *labdomain.LabReport) error {
	var documentID pgtype.UUID
	err := r.client.Pool().QueryRow(ctx,
		`SELECT exam_document_id FROM lab_reports WHERE id = $1 AND patient_id = $2`,
		report.ID, report.PatientID,
	).Scan(&documentID)
	if err != nil {
		return err
	}
	if documentID.Valid {
		id := uuid.UUID(documentID.Bytes)
		report.ExamDocumentID = &id
	}
	return nil
}

func mapProcessingWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "idx_lab_reports_fingerprint", "idx_lab_reports_exam_document":
			return errors.Join(labdomain.ErrLabReportAlreadyExists, err)
		}
	}
	return err
}
