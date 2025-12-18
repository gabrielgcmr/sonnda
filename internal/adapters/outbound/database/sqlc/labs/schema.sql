CREATE TABLE lab_reports (
    id                  uuid PRIMARY KEY,
    patient_id          uuid NOT NULL,
    patient_name        text,
    patient_dob         date,
    lab_name            text,
    lab_phone           text,
    insurance_provider  text,
    requesting_doctor   text,
    technical_manager   text,
    report_date         date,
    raw_text            text,
    uploaded_by_user_id uuid,
    fingerprint         text,
    created_at          timestamp with time zone NOT NULL,
    updated_at          timestamp with time zone NOT NULL
);

CREATE TABLE lab_results (
    id            uuid PRIMARY KEY,
    lab_report_id uuid NOT NULL REFERENCES lab_reports(id),
    test_name     text NOT NULL,
    material      text,
    method        text,
    collected_at  timestamp with time zone,
    release_at    timestamp with time zone
);

CREATE TABLE lab_result_items (
    id                uuid PRIMARY KEY,
    lab_result_id uuid NOT NULL REFERENCES lab_results(id),
    parameter_name    text NOT NULL,
    result_value       text,
    result_unit       text,
    reference_text    text
);