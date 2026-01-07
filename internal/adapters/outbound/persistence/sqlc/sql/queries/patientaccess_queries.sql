-- internal/adapters/outbound/database/sqlc/patientaccess/queries.sql

-- name: UpsertPatientAccess :exec
INSERT INTO patient_access (
    patient_id,
    user_id,
    role
) VALUES ($1, $2, $3)
ON CONFLICT (patient_id, user_id)
DO UPDATE SET
    role = EXCLUDED.role,
    updated_at = now();

-- name: FindPatientAccess :one
SELECT
    patient_id,
    user_id,
    role,
    created_at,
    updated_at
FROM patient_access
WHERE patient_id = $1
  AND user_id = $2;

-- name: ListPatientAccessByPatient :many
SELECT
    patient_id,
    user_id,
    role,
    created_at,
    updated_at
FROM patient_access
WHERE patient_id = $1
ORDER BY user_id;

-- name: ListPatientAccessByUser :many
SELECT
    patient_id,
    user_id,
    role,
    created_at,
    updated_at
FROM patient_access
WHERE user_id = $1
ORDER BY patient_id;

-- name: RevokePatientAccess :execrows
DELETE FROM patient_access
WHERE patient_id = $1
  AND user_id = $2;
