-- name: CreateUser :one
INSERT INTO users (
  id, auth_provider, auth_subject, email, full_name, birth_date, cpf, phone, role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: FindUserByAuthIdentity :one
SELECT *
FROM users
WHERE auth_provider = $1
  AND auth_subject = $2
  AND deleted_at IS NULL;

-- name: FindUserByEmail :one
SELECT *
FROM users
WHERE email = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: FindUserByID :one
SELECT *
FROM users
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: FindUserByCPF :one
SELECT *
FROM users
WHERE cpf = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: SoftDeleteUser :execrows
UPDATE users
SET deleted_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

--Profissionais

-- name: CreateProfessional :one
-- Cria apenas a parte "profissional" (O ID vem do User já criado)
INSERT INTO professionals (
  user_id, registration_number, registration_issuer, registration_state, status
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetProfessionalByUserID :one
SELECT * 
FROM professionals
WHERE user_id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListProfessionalsByName :many
-- AQUI ESTÁ O TRUQUE: Fazemos JOIN para filtrar, mas retornamos dados do profissional
SELECT p.*
FROM professionals p
JOIN users u ON u.id = p.user_id
WHERE u.full_name ILIKE '%' || sqlc.arg(name) || '%'
  AND p.deleted_at IS NULL
LIMIT $1 OFFSET $2;

-- name: GetFullProfessionalDetails :one
-- Query especial para telas de perfil: Retorna TUDO junto
SELECT 
    p.*,
    u.full_name,
    u.email,
    u.phone
FROM professionals p
JOIN users u ON u.id = p.user_id
WHERE p.user_id = $1 AND p.deleted_at IS NULL
LIMIT 1;
