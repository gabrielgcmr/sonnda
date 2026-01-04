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

-- 2. Tabela de Profissionais
-- Reflete a entidade 'Professional' que criamos.
-- Relacionamento 1:1 com users.
CREATE TABLE professionals (
    user_id             UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    kind                TEXT NOT NULL CHECK (kind IN ('doctor','nurse','nursing_tech','physiotherapist','psychologist','nutritionist','pharmacist','dentist')),
    registration_number TEXT NOT NULL,
    registration_issuer TEXT NOT NULL, -- Ex: CRM, COREN
    registration_state  TEXT,          -- Opcional (UF)
    status              TEXT NOT NULL CHECK (status IN ('pending', 'verified', 'rejected')),
    verified_at         TIMESTAMP WITH TIME ZONE,
    deleted_at          TIMESTAMP WITH TIME ZONE,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Índices para profissionais
CREATE INDEX idx_professionals_status ON professionals(status);
CREATE INDEX idx_professionals_reg_number ON professionals(registration_number);
CREATE INDEX idx_professionals_deleted_at ON professionals (deleted_at);

