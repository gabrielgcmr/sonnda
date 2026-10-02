-- internal/infrastructure/persistence/postgres/sqlc/sql/queries/exam_queries.sql
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

-- name: ListExamDocumentTextsByPatientID :many
SELECT *
FROM exam_document_texts
WHERE patient_id = $1
ORDER BY performed_at DESC NULLS LAST, created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetExamDocumentExtraction :one
SELECT snapshot FROM exam_document_extractions WHERE document_id = $1;