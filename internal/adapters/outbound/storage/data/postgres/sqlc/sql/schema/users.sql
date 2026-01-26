-- Users table: stores core user information
CREATE TABLE users (
    id            UUID PRIMARY KEY,
    auth_provider TEXT NOT NULL,
    auth_subject  TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    full_name     TEXT NOT NULL,
    birth_date    DATE NOT NULL,
    cpf           TEXT NOT NULL UNIQUE,
    phone         TEXT NOT NULL,
    account_type  TEXT NOT NULL CHECK (account_type IN ('professional', 'basic_care')),
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMP WITH TIME ZONE
);

-- Índices de busca rápida e unicidade de identidade
CREATE UNIQUE INDEX idx_users_auth_identity ON users(auth_provider, auth_subject);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_cpf ON users(cpf);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

