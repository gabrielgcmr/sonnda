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

-- name: CreateLabPanel :one
INSERT INTO lab_panels(
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

-- name: CreateObservation :one
INSERT INTO observations (
    id,
  lab_panel_id,
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

-- name: GetLabPanelsByReportID :one
SELECT
    id,
    test_name,
    material,
    method,
    collected_at,
    release_at
FROM lab_panels
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

-- name: ListLabPanelsByReportID :many
SELECT
  id, lab_report_id, test_name, material, method, collected_at, release_at
FROM lab_panels
WHERE lab_report_id = $1
ORDER BY collected_at NULLS LAST, id;

-- name: ListObservationsByPanelID :many
SELECT
  id, lab_panel_id, parameter_name, result_value, result_unit, reference_text
FROM observations
WHERE lab_panel_id = $1
ORDER BY id;

-- ============================================================
-- Timeline
-- ============================================================

-- name: ListObservationTimelineByPatientAndParameter :many
SELECT
  lr.id AS report_id,
  p.id AS lab_panel_id,
  o.id AS observation_id,
  lr.report_date AS report_date,
  p.test_name AS test_name,
  o.parameter_name,
  o.result_value,
  o.result_unit,
  o.reference_text
FROM observations o
JOIN lab_panels p ON o.lab_panel_id = p.id
JOIN lab_reports lr ON p.lab_report_id = lr.id
WHERE lr.patient_id = $1
  AND o.parameter_name = $2
ORDER BY lr.report_date DESC NULLS LAST, lr.created_at DESC
LIMIT $3 OFFSET $4;

-- ============================================================
-- Deletes
-- ============================================================

-- name: DeleteLabReport :execrows
DELETE FROM lab_reports
WHERE id = $1;