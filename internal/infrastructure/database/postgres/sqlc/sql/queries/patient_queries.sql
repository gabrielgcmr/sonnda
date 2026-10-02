-- internal/adapters/outbound/database/sqlc/patients/queries.sql

-- Common column set for patient fetches:
-- id, owner_user_id, cpf, cns, full_name, birth_date, gender, race, phone, avatar_url, created_at, updated_at

-- name: CreatePatient :one
INSERT INTO patients (
    id,
    owner_user_id,
    cpf,
    cns,
    full_name,
    birth_date,
    gender,
    race,
    phone,
    avatar_url,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    now(), now()
)
RETURNING *;

-- name: GetPatientByID :one
SELECT *
FROM patients
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetPatientByCPF :one
SELECT *
FROM patients
WHERE cpf = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetPatientByCNS :one
SELECT *
FROM patients
WHERE cns = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetPatientByOwnerUserID :one
SELECT *
FROM patients
WHERE owner_user_id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: ListPatients :many
SELECT *
FROM patients
WHERE deleted_at IS NULL
ORDER BY full_name
LIMIT $1 OFFSET $2;

-- name: SearchPatientsByName :many
SELECT *
FROM patients
WHERE deleted_at IS NULL
  AND full_name ILIKE '%' || sqlc.arg(query) || '%'
ORDER BY full_name
LIMIT $1 OFFSET $2;

-- name: UpdatePatient :one
UPDATE patients
SET
    full_name  = COALESCE($2, full_name),
    phone      = COALESCE($3, phone),
    avatar_url = COALESCE($4, avatar_url),
    gender     = COALESCE($5, gender),
    race       = COALESCE($6, race),
    cns        = COALESCE($7, cns),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePatient :execrows
UPDATE patients
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: RestorePatient :execrows
UPDATE patients
SET deleted_at = NULL,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NOT NULL;
