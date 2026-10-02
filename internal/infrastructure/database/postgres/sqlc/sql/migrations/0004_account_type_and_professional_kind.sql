-- Introduce `users.account_type` and `professionals.kind`.
-- This migration is safe to apply after `0002_create_users.sql`.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'role'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'account_type'
    ) THEN
        ALTER TABLE users RENAME COLUMN role TO account_type;
    END IF;
END
$$;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_account_type_check;
ALTER TABLE users
    ADD CONSTRAINT users_account_type_check CHECK (account_type IN ('professional', 'basic_care'));

UPDATE users
SET account_type = 'basic_care'
WHERE account_type = 'caregiver';

ALTER TABLE professionals ADD COLUMN IF NOT EXISTS kind TEXT;
ALTER TABLE professionals DROP CONSTRAINT IF EXISTS professionals_kind_check;
ALTER TABLE professionals
    ADD CONSTRAINT professionals_kind_check CHECK (kind IN ('doctor','nurse','nursing_tech','physiotherapist','psychologist','nutritionist','pharmacist','dentist'));

UPDATE professionals
SET kind = 'doctor'
WHERE kind IS NULL;

ALTER TABLE professionals ALTER COLUMN kind SET NOT NULL;

