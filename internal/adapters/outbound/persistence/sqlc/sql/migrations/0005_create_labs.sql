-- +migrate Up
-- Lab reports: optional extracted metadata, linked to patient and uploader.
CREATE TABLE lab_reports (
    id                 UUID PRIMARY KEY,
    patient_id         UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    uploaded_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    patient_name       TEXT,
    patient_dob        TIMESTAMP WITH TIME ZONE,
    lab_name           TEXT,
    lab_phone          TEXT,
    insurance_provider TEXT,
    requesting_doctor  TEXT,
    technical_manager  TEXT,
    report_date        TIMESTAMP WITH TIME ZONE,
    raw_text           TEXT,
    fingerprint        TEXT,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Lab results: one-to-many from lab_reports.
CREATE TABLE lab_results (
    id            UUID PRIMARY KEY,
    lab_report_id UUID   NOT NULL REFERENCES lab_reports(id) ON DELETE CASCADE,
    test_name     TEXT NOT NULL,
    material      TEXT,
    method        TEXT,
    collected_at  TIMESTAMP WITH TIME ZONE,
    release_at    TIMESTAMP WITH TIME ZONE
);

-- Lab result items: one-to-many from lab_results.
CREATE TABLE lab_result_items (
    id             UUID PRIMARY KEY,
    lab_result_id  UUID NOT NULL REFERENCES lab_results(id) ON DELETE CASCADE,
    parameter_name TEXT NOT NULL,
    result_value   TEXT,
    result_unit    TEXT,
    reference_text TEXT
);

-- Useful indexes/uniqueness for lookups and idempotency
CREATE UNIQUE INDEX idx_lab_reports_fingerprint ON lab_reports(fingerprint) WHERE fingerprint IS NOT NULL;
CREATE INDEX idx_lab_reports_patient ON lab_reports(patient_id);
CREATE INDEX idx_lab_reports_report_date ON lab_reports(report_date);
CREATE INDEX idx_lab_results_report ON lab_results(lab_report_id);
CREATE INDEX idx_lab_result_items_result ON lab_result_items(lab_result_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_lab_result_items_result;
DROP INDEX IF EXISTS idx_lab_results_report;
DROP INDEX IF EXISTS idx_lab_reports_report_date;
DROP INDEX IF EXISTS idx_lab_reports_patient;
DROP INDEX IF EXISTS idx_lab_reports_fingerprint;
DROP TABLE IF EXISTS lab_result_items;
DROP TABLE IF EXISTS lab_results;
DROP TABLE IF EXISTS lab_reports;
