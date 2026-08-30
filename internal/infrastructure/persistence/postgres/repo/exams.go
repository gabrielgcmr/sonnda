package repo

import (
	"context"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	postgress "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	examsqlc "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres/sqlc/generated/exam"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ExamsRepository struct {
	client  *postgress.Client
	queries *examsqlc.Queries
}

var _ repository.Exams = (*ExamsRepository)(nil)

func NewExamsRepository(client *postgress.Client) repository.Exams {
	return &ExamsRepository{
		client:  client,
		queries: examsqlc.New(client.Pool()),
	}
}

func (r *ExamsRepository) Create(ctx context.Context, document *exams.ExamDocument) error {
	if document == nil {
		return ErrRepositoryFailure
	}
	if err := document.NormalizeAndValidate(); err != nil {
		return err
	}

	row, err := r.queries.CreateExamDocument(ctx, examsqlc.CreateExamDocumentParams{
		ID:               document.ID,
		PatientID:        document.PatientID,
		UploadedByUserID: document.UploadedByUserID,
		StorageUri:       document.StorageURI,
		OriginalFilename: document.OriginalFilename,
		MimeType:         document.MimeType,
		Status:           string(document.Status),
		ExamType:         examTypeToPgText(document.ExamType),
		ExtractionMethod: FromNullableStringToPgText(document.ExtractionMethod),
		Confidence:       nullableFloat64ToPgFloat8(document.Confidence),
		ExtractedText:    FromNullableStringToPgText(document.ExtractedText),
		ErrorMessage:     FromNullableStringToPgText(document.ErrorMessage),
		CreatedAt:        FromRequiredTimestamptzToPgTimestamptz(document.CreatedAt),
		UpdatedAt:        FromRequiredTimestamptzToPgTimestamptz(document.UpdatedAt),
	})
	if err != nil {
		return err
	}

	*document = mapExamDocumentRow(row)
	return nil
}

func (r *ExamsRepository) FindByID(ctx context.Context, id uuid.UUID) (*exams.ExamDocument, error) {
	row, err := r.queries.GetExamDocumentByID(ctx, id)
	if err != nil {
		if IsPgNotFound(err) {
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

func (r *ExamsRepository) MarkProcessing(ctx context.Context, id uuid.UUID, extractionMethod *string) (*exams.ExamDocument, error) {
	row, err := r.queries.UpdateExamDocumentProcessing(ctx, examsqlc.UpdateExamDocumentProcessingParams{
		ID:               id,
		ExtractionMethod: FromNullableStringToPgText(extractionMethod),
		UpdatedAt:        FromRequiredTimestamptzToPgTimestamptz(time.Now().UTC()),
	})
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	document := mapExamDocumentRow(row)
	return &document, nil
}

func (r *ExamsRepository) MarkClassified(
	ctx context.Context,
	id uuid.UUID,
	status exams.DocumentStatus,
	examType exams.ExamType,
	extractionMethod *string,
	confidence *float64,
	extractedText *string,
) (*exams.ExamDocument, error) {
	row, err := r.queries.UpdateExamDocumentClassified(ctx, examsqlc.UpdateExamDocumentClassifiedParams{
		ID:               id,
		Status:           string(status),
		ExamType:         examTypeToPgText(&examType),
		ExtractionMethod: FromNullableStringToPgText(extractionMethod),
		Confidence:       nullableFloat64ToPgFloat8(confidence),
		ExtractedText:    FromNullableStringToPgText(extractedText),
		UpdatedAt:        FromRequiredTimestamptzToPgTimestamptz(time.Now().UTC()),
	})
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	document := mapExamDocumentRow(row)
	return &document, nil
}

func (r *ExamsRepository) MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) (*exams.ExamDocument, error) {
	row, err := r.queries.UpdateExamDocumentFailed(ctx, examsqlc.UpdateExamDocumentFailedParams{
		ID:           id,
		ErrorMessage: FromRequiredStringToPgText(errorMessage),
		UpdatedAt:    FromRequiredTimestamptzToPgTimestamptz(time.Now().UTC()),
	})
	if err != nil {
		if IsPgNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	document := mapExamDocumentRow(row)
	return &document, nil
}

func mapExamDocumentRow(row examsqlc.ExamDocument) exams.ExamDocument {
	return exams.ExamDocument{
		ID:               row.ID,
		PatientID:        row.PatientID,
		UploadedByUserID: row.UploadedByUserID,
		StorageURI:       row.StorageUri,
		OriginalFilename: row.OriginalFilename,
		MimeType:         row.MimeType,
		Status:           exams.DocumentStatus(row.Status),
		ExamType:         pgTextToExamType(row.ExamType),
		ExtractionMethod: FromPgTextToNullableString(row.ExtractionMethod),
		Confidence:       pgFloat8ToNullableFloat64(row.Confidence),
		ExtractedText:    FromPgTextToNullableString(row.ExtractedText),
		ErrorMessage:     FromPgTextToNullableString(row.ErrorMessage),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func examTypeToPgText(examType *exams.ExamType) pgtype.Text {
	if examType == nil {
		return pgtype.Text{Valid: false}
	}
	return FromRequiredStringToPgText(string(*examType))
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
