-- internal/adapters/outbound/database/sqlc/patientaccess/queries.sql

-- name: UpsertPatientAccess :exec
INSERT INTO patient_access (
    patient_id,
    grantee_id,
    relation_type,
    granted_by
) VALUES ($1, $2, $3, $4)
ON CONFLICT (patient_id, grantee_id)
DO UPDATE SET
    relation_type = EXCLUDED.relation_type,
    granted_by = EXCLUDED.granted_by,
    revoked_at = NULL;

-- name: FindPatientAccess :one
SELECT
    patient_id,
    grantee_id,
    relation_type,
    created_at,
    revoked_at,
    granted_by
FROM patient_access
WHERE patient_id = $1
  AND grantee_id = $2;

-- name: ListPatientAccessByPatient :many
SELECT
    patient_id,
    grantee_id,
    relation_type,
    created_at,
    revoked_at,
    granted_by
FROM patient_access
WHERE patient_id = $1
ORDER BY grantee_id;

-- name: ListPatientAccessByUser :many
SELECT
    patient_id,
    grantee_id,
    relation_type,
    created_at,
    revoked_at,
    granted_by
FROM patient_access
WHERE grantee_id = $1
ORDER BY patient_id;

-- name: RevokePatientAccess :execrows
UPDATE patient_access
SET revoked_at = now()
WHERE patient_id = $1
  AND grantee_id = $2
  AND revoked_at IS NULL;

-- Minimal list of patients accessible by a user (for UI listing)
-- Returns patient basic info and the relation type. Paginates by full_name.
-- name: ListAccessiblePatientsByUser :many
SELECT
    pa.patient_id,
    p.full_name,
    p.avatar_url,
    pa.relation_type
FROM patient_access pa
JOIN patients p ON p.id = pa.patient_id
WHERE pa.grantee_id = $1
  AND pa.revoked_at IS NULL
  AND p.deleted_at IS NULL
ORDER BY p.full_name
LIMIT $2 OFFSET $3;

-- Total count for pagination of accessible patients by user
-- name: CountAccessiblePatientsByUser :one
SELECT COUNT(*) AS total
FROM patient_access pa
JOIN patients p ON p.id = pa.patient_id
WHERE pa.grantee_id = $1
  AND pa.revoked_at IS NULL
  AND p.deleted_at IS NULL;
