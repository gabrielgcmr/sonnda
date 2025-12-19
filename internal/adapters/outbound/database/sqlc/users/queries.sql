-- internal/adapters/outbound/database/sqlc/users/queries.sql

-- name: FindUserByAuthIdentity :one
SELECT
    id,
    auth_provider,
    auth_subject,
    email,
    role,
    created_at,
    updated_at
FROM app_users
WHERE auth_provider = $1
  AND auth_subject = $2;

-- name: FindUserByEmail :one
SELECT
    id,
    auth_provider,
    auth_subject,
    email,
    role,
    created_at,
    updated_at
FROM app_users
WHERE email = $1;

-- name: FindUserByID :one
SELECT
    id,
    auth_provider,
    auth_subject,
    email,
    role,
    created_at,
    updated_at
FROM app_users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO app_users (auth_provider, auth_subject, email, role)
VALUES ($1, $2, $3, $4)
RETURNING
    id,
    auth_provider,
    auth_subject,
    email,
    role,
    created_at,
    updated_at;

-- name: UpdateUserAuthIdentity :one
UPDATE app_users
SET auth_provider = $1,
    auth_subject  = $2,
    updated_at    = now()
WHERE id = $3
RETURNING
    id,
    auth_provider,
    auth_subject,
    email,
    role,
    created_at,
    updated_at;
