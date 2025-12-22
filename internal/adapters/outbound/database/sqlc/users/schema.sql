CREATE TABLE app_users (
    id            TEXT PRIMARY KEY,
    auth_provider text NOT NULL,
    auth_subject  text NOT NULL,
    email         text NOT NULL,
    role          text NOT NULL,
    created_at    timestamp with time zone NOT NULL DEFAULT now(),
    updated_at    timestamp with time zone NOT NULL DEFAULT now()
);
