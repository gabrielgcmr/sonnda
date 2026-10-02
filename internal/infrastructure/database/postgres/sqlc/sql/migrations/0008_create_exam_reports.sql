-- +migrate Up
CREATE TABLE IF NOT EXISTS exam_reports (
    id                  UUID PRIMARY KEY,
    exam_document_id    UUID UNIQUE REFERENCES exam_documents(id) ON DELETE SET NULL,
    patient_id          UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    uploaded_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category            TEXT NOT NULL,
    title               TEXT,
    modality            TEXT,
    body_site           TEXT,
    performed_at        TIMESTAMP WITH TIME ZONE,
    facility_name       TEXT,
    interpreting_doctor TEXT,
    report_text         TEXT NOT NULL,
    conclusion          TEXT,
    extraction_method   TEXT,
    confidence          DOUBLE PRECISION,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    CONSTRAINT chk_exam_reports_category CHECK (
        category IN ('laboratory', 'imaging', 'unknown')
    ),
    CONSTRAINT chk_exam_reports_confidence CHECK (
        confidence IS NULL OR (confidence >= 0 AND confidence <= 1)
    )
);

CREATE INDEX IF NOT EXISTS idx_exam_reports_patient ON exam_reports(patient_id);
CREATE INDEX IF NOT EXISTS idx_exam_reports_patient_created_at ON exam_reports(patient_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_exam_reports_category ON exam_reports(category);

-- +migrate Down
DROP INDEX IF EXISTS idx_exam_reports_category;
DROP INDEX IF EXISTS idx_exam_reports_patient_created_at;
DROP INDEX IF EXISTS idx_exam_reports_patient;
DROP TABLE IF EXISTS exam_reports;
