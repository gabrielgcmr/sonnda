CREATE TABLE app_users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    auth_provider text NOT NULL,
    auth_subject  text NOT NULL,
    email         text NOT NULL,
    role          text NOT NULL,
    created_at    timestamp with time zone NOT NULL DEFAULT now(),
    updated_at    timestamp with time zone NOT NULL DEFAULT now()
);
