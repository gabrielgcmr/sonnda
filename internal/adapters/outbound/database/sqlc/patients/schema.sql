CREATE TABLE patients (
    id           TEXT PRIMARY KEY,
    app_user_id  text REFERENCES app_users(id),
    cpf          text NOT NULL,
    cns          text,
    full_name    text NOT NULL,
    birth_date   date NOT NULL,
    gender       text NOT NULL,
    race         text NOT NULL,
    phone        text,
    avatar_url   text NOT NULL,
    created_at   timestamp with time zone NOT NULL DEFAULT now(),
    updated_at   timestamp with time zone NOT NULL DEFAULT now()
);
