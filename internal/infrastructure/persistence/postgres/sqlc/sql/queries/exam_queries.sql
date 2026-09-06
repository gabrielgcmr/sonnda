-- internal/infrastructure/persistence/postgres/sqlc/sql/queries/exam_queries.sql
-- name: CreateExamDocument :one
INSERT INTO exam_documents (
    id,
    patient_id,
    uploaded_by_user_id,
    storage_uri,
    original_filename,
    mime_type,
    status,
    exam_type,
    extraction_method,
    confidence,
    extracted_text,
    error_message,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: GetExamDocumentByID :one
SELECT *
FROM exam_documents
WHERE id = $1;

-- name: ListExamDocumentsByPatientID :many
SELECT *
FROM exam_documents
WHERE patient_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateExamDocumentProcessing :one
UPDATE exam_documents
SET
    status = 'processing',
    extraction_method = $2,
    error_message = NULL,
    updated_at = $3
WHERE id = $1
RETURNING *;

-- name: UpdateExamDocumentClassified :one
UPDATE exam_documents
SET
    status = $2,
    exam_type = $3,
    extraction_method = $4,
    confidence = $5,
    extracted_text = $6,
    error_message = $8,
    updated_at = $7
WHERE id = $1
RETURNING *;

-- name: UpdateExamDocumentFailed :one
UPDATE exam_documents
SET
    status = 'failed',
    error_message = $2,
    updated_at = $3
WHERE id = $1
RETURNING *;

-- name: CreateExamDocumentText :one
INSERT INTO exam_document_texts (
    id,
    exam_document_id,
    patient_id,
    uploaded_by_user_id,
    category,
    title,
    modality,
    body_site,
    performed_at,
    facility_name,
    interpreting_doctor,
    text,
    conclusion,
    extraction_method,
    confidence,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING *;

-- name: GetExamDocumentTextByID :one
SELECT *
FROM exam_document_texts
WHERE id = $1;

-- name: GetExamDocumentTextByDocumentID :one
SELECT *
FROM exam_document_texts
WHERE exam_document_id = $1;

-- name: ListExamDocumentTextsByPatientID :many
SELECT *
FROM exam_document_texts
WHERE patient_id = $1
ORDER BY performed_at DESC NULLS LAST, created_at DESC
LIMIT $2 OFFSET $3;
