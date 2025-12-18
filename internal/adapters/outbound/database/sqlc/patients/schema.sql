CREATE TABLE patients (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    app_user_id  uuid REFERENCES app_users(id),
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
