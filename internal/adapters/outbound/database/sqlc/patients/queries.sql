-- internal/adapters/outbound/database/sqlc/patients/queries.sql

-- Common column set for patient fetches:
-- id, app_user_id, cpf, cns, full_name, birth_date, gender, race, phone, avatar_url, created_at, updated_at

-- name: CreatePatient :one
INSERT INTO patients (
    app_user_id,
    cpf,
    cns,
    full_name,
    birth_date,
    gender,
    race,
    phone,
    avatar_url
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING
    id,
    app_user_id,
    cpf,
    cns,
    full_name,
    birth_date,
    gender,
    race,
    phone,
    avatar_url,
    created_at,
    updated_at;

-- name: UpdatePatient :one
UPDATE patients
SET cpf        = $2,
    cns        = $3,
    full_name  = $4,
    birth_date = $5,
    gender     = $6,
    race       = $7,
    phone      = $8,
    avatar_url = $9,
    updated_at = now()
WHERE id = $1
RETURNING
    id,
    app_user_id,
    cpf,
    cns,
    full_name,
    birth_date,
    gender,
    race,
    phone,
    avatar_url,
    created_at,
    updated_at;

-- name: DeletePatient :execrows
DELETE FROM patients WHERE id = $1;

-- name: FindPatientByUserID :one
SELECT
    id,
    app_user_id,
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
FROM patients
WHERE app_user_id = $1
LIMIT 1;

-- name: FindPatientByCPF :one
SELECT
    id,
    app_user_id,
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
FROM patients
WHERE cpf = $1
LIMIT 1;

-- name: FindPatientByID :one
SELECT
    id,
    app_user_id,
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
FROM patients
WHERE id = $1
LIMIT 1;

-- name: ListPatients :many
SELECT
    id,
    app_user_id,
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
FROM patients
ORDER BY full_name ASC
LIMIT $1 OFFSET $2;
