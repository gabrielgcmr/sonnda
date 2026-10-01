// internal/features/documentprocessing/postgres/document_repository.go
package postgres

import (
	"context"

	processing "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	exams "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	examsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/exam"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ExamsRepository struct {
	client  *postgress.Client
	queries *examsqlc.Queries
}

var _ processing.DocumentRepository = (*ExamsRepository)(nil)

func NewDocumentRepository(client *postgress.Client) processing.DocumentRepository {
	return &ExamsRepository{
		client:  client,
		queries: examsqlc.New(client.Pool()),
	}
}

func (r *ExamsRepository) FindByID(ctx context.Context, id uuid.UUID) (*exams.ExamDocument, error) {
	row, err := r.queries.GetExamDocumentByID(ctx, id)
	if err != nil {
		if isDocumentNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	document := mapExamDocumentRow(row)
	return &document, nil
}

func (r *ExamsRepository) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]exams.ExamDocument, error) {
	rows, err := r.queries.ListExamDocumentsByPatientID(ctx, examsqlc.ListExamDocumentsByPatientIDParams{
		PatientID: patientID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}

	documents := make([]exams.ExamDocument, 0, len(rows))
	for _, row := range rows {
		documents = append(documents, mapExamDocumentRow(row))
	}

	return documents, nil
}

func (r *ExamsRepository) ListDocumentTextsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]exams.ExamDocumentText, error) {
	rows, err := r.queries.ListExamDocumentTextsByPatientID(ctx, examsqlc.ListExamDocumentTextsByPatientIDParams{
		PatientID: patientID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}

	documentTexts := make([]exams.ExamDocumentText, 0, len(rows))
	for _, row := range rows {
		documentTexts = append(documentTexts, mapExamDocumentTextRow(row))
	}

	return documentTexts, nil
}

func mapExamDocumentRow(row examsqlc.ExamDocument) exams.ExamDocument {
	return exams.ExamDocument{
		ReviewStatus: pgTextToNullableString(row.ReviewStatus), LabReportID: pgUUIDToNullableUUID(row.LabReportID), ConfirmedByUserID: pgUUIDToNullableUUID(row.ConfirmedByUserID), ConfirmedAt: pgTimestamptzToNullableTime(row.ConfirmedAt),
		ID:               row.ID,
		PatientID:        row.PatientID,
		UploadedByUserID: row.UploadedByUserID,
		StorageURI:       row.StorageUri,
		OriginalFilename: row.OriginalFilename,
		MimeType:         row.MimeType,
		Status:           exams.DocumentStatus(row.Status),
		ExamType:         pgTextToExamType(row.ExamType),
		ExtractionMethod: pgTextToNullableString(row.ExtractionMethod),
		Confidence:       pgFloat8ToNullableFloat64(row.Confidence),
		ExtractedText:    pgTextToNullableString(row.ExtractedText),
		ErrorMessage:     pgTextToNullableString(row.ErrorMessage),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func mapExamDocumentTextRow(row examsqlc.ExamDocumentText) exams.ExamDocumentText {
	return exams.ExamDocumentText{
		ID:                 row.ID,
		ExamDocumentID:     pgUUIDToNullableUUID(row.ExamDocumentID),
		PatientID:          row.PatientID,
		UploadedByUserID:   row.UploadedByUserID,
		Category:           exams.ExamType(row.Category),
		Title:              pgTextToNullableString(row.Title),
		Modality:           pgTextToNullableString(row.Modality),
		BodySite:           pgTextToNullableString(row.BodySite),
		PerformedAt:        pgTimestamptzToNullableTime(row.PerformedAt),
		FacilityName:       pgTextToNullableString(row.FacilityName),
		InterpretingDoctor: pgTextToNullableString(row.InterpretingDoctor),
		Text:               row.Text,
		Conclusion:         pgTextToNullableString(row.Conclusion),
		ExtractionMethod:   pgTextToNullableString(row.ExtractionMethod),
		Confidence:         pgFloat8ToNullableFloat64(row.Confidence),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}

func examTypeToPgText(examType *exams.ExamType) pgtype.Text {
	if examType == nil {
		return pgtype.Text{Valid: false}
	}
	return requiredStringToPgText(string(*examType))
}

func pgTextToExamType(value pgtype.Text) *exams.ExamType {
	if !value.Valid {
		return nil
	}
	examType := exams.ExamType(value.String)
	return &examType
}

func nullableFloat64ToPgFloat8(value *float64) pgtype.Float8 {
	if value == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *value, Valid: true}
}

func pgFloat8ToNullableFloat64(value pgtype.Float8) *float64 {
	if !value.Valid {
		return nil
	}
	result := value.Float64
	return &result
}
