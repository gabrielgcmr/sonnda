-- internal/infrastructure/persistence/postgres/sqlc/sql/queries/lab_queries.sql

-- name: LabDocumentBelongsToPatient :one
SELECT EXISTS (
    SELECT 1 FROM exam_documents WHERE id = $1 AND patient_id = $2
);

-- name: AttachLabReportDocument :execrows
UPDATE lab_reports AS l
SET exam_document_id = d.id, updated_at = now()
FROM exam_documents AS d
WHERE l.id = sqlc.arg(report_id)
  AND l.patient_id = sqlc.arg(patient_id)
  AND d.id = sqlc.arg(document_id)
  AND d.patient_id = l.patient_id
  AND (l.exam_document_id IS NULL OR l.exam_document_id = d.id);

-- ============================================================
-- Creators
-- ============================================================

-- name: CreateLabReport :one
INSERT INTO lab_reports (
    id,
    patient_id,
    exam_document_id,
    patient_name,
    patient_dob,
    lab_name,
    lab_phone,
    insurance_provider,
    requesting_doctor,
    technical_manager,
    report_date,
    raw_text,
    uploaded_by_user_id,
    fingerprint
)
VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12,
    $13, $14
)
RETURNING
    id,
    patient_id,
    exam_document_id,
    patient_name,
    patient_dob,
    lab_name,
    lab_phone,
    insurance_provider,
    requesting_doctor,
    technical_manager,
    report_date,
    raw_text,
    uploaded_by_user_id,
    fingerprint,
    created_at,
    updated_at;

-- name: CreateLabResult :one
INSERT INTO lab_results(
    id,
    lab_report_id,
    test_name,
    material,
    method,
    collected_at,
    release_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING id;

-- name: CreateLabResultItem :one
INSERT INTO lab_result_items (
    id,
    lab_result_id,
    parameter_name,
    result_value,
    result_unit,
    reference_text
)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id;

-- ============================================================
-- Getters
-- ============================================================

-- name: GetLabReportByID :one
SELECT
    id,
    patient_id,
    exam_document_id,
    patient_name,
    patient_dob,
    lab_name,
    lab_phone,
    insurance_provider,
    requesting_doctor,
    technical_manager,
    report_date,
    raw_text,
    uploaded_by_user_id,
    fingerprint,
    created_at,
    updated_at
FROM lab_reports
WHERE id = $1;

-- name: GetLabResultsByReportID :one
SELECT
    id,
    test_name,
    material,
    method,
    collected_at,
    release_at
FROM lab_results
WHERE lab_report_id = $1
ORDER BY test_name;

-- ============================================================
-- Dedupe (Existence checks)
-- ============================================================

-- name: ExistsLabReportByPatientAndFingerprint :one
SELECT EXISTS(
  SELECT 1
  FROM lab_reports
  WHERE patient_id  = $1
    AND fingerprint = $2
);

-- name: GetLabReportByPatientAndFingerprint :one
SELECT
    id,
    patient_id,
    exam_document_id,
    patient_name,
    patient_dob,
    lab_name,
    lab_phone,
    insurance_provider,
    requesting_doctor,
    technical_manager,
    report_date,
    raw_text,
    uploaded_by_user_id,
    fingerprint,
    created_at,
    updated_at
FROM lab_reports
WHERE patient_id = $1
  AND fingerprint = $2;

-- ============================================================
-- List
-- ============================================================

-- name: ListLabReportsByPatientID :many
SELECT
    id,
    patient_id,
    exam_document_id,
    patient_name,
    lab_name,
    report_date,
    uploaded_by_user_id,
    fingerprint,
    created_at,
    updated_at
FROM lab_reports
WHERE patient_id = $1
ORDER BY report_date DESC NULLS LAST, created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListLabResultsByReportID :many
SELECT
  id, lab_report_id, test_name, material, method, collected_at, release_at
FROM lab_results
WHERE lab_report_id = $1
ORDER BY collected_at NULLS LAST, id;

-- name: ListLabResultItemsByResultID :many
SELECT
  id, lab_result_id, parameter_name, result_value, result_unit, reference_text
FROM lab_result_items
WHERE lab_result_id = $1
ORDER BY id;

-- ============================================================
-- Timeline
-- ============================================================

-- name: ListLabItemTimelineByPatientAndParameter :many
SELECT
  lr.id           AS report_id,
  r.id          AS lab_result_id,
  i.id          AS item_id,
  lr.report_date  AS report_date,
  r.test_name   AS test_name,
  i.parameter_name,
  i.result_value,
  i.result_unit
FROM lab_result_items i
JOIN lab_results r ON i.lab_result_id = r.id
JOIN lab_reports      lr  ON r.lab_report_id      = lr.id
WHERE lr.patient_id      = $1
  AND i.parameter_name = $2
ORDER BY lr.report_date DESC NULLS LAST, lr.created_at DESC
LIMIT $3 OFFSET $4;

-- ============================================================
-- Deletes
-- ============================================================

-- name: DeleteLabResultItemsByReportID :execrows
DELETE FROM lab_result_items
WHERE lab_result_id IN (
  SELECT id FROM lab_results WHERE lab_report_id = $1
);

-- name: DeleteLabResultsByReportID :execrows
DELETE FROM lab_results
WHERE lab_report_id = $1;

-- name: DeleteLabReport :execrows
DELETE FROM lab_reports
WHERE id = $1;
