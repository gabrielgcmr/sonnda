-- +migrate Up
CREATE TABLE exam_documents (
    id                  UUID PRIMARY KEY,
    patient_id          UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    uploaded_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    storage_uri         TEXT NOT NULL,
    original_filename   TEXT NOT NULL,
    mime_type           TEXT NOT NULL,
    status              TEXT NOT NULL,
    exam_type           TEXT,
    extraction_method   TEXT,
    confidence          DOUBLE PRECISION,
    extracted_text      TEXT,
    error_message       TEXT,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    CONSTRAINT chk_exam_documents_status CHECK (
        status IN ('uploaded', 'processing', 'processed', 'failed', 'needs_review')
    ),
    CONSTRAINT chk_exam_documents_exam_type CHECK (
        exam_type IS NULL OR exam_type IN ('laboratory', 'imaging', 'unknown')
    ),
    CONSTRAINT chk_exam_documents_confidence CHECK (
        confidence IS NULL OR (confidence >= 0 AND confidence <= 1)
    )
);

CREATE INDEX idx_exam_documents_patient ON exam_documents(patient_id);
CREATE INDEX idx_exam_documents_patient_created_at ON exam_documents(patient_id, created_at DESC);
CREATE INDEX idx_exam_documents_status ON exam_documents(status);
CREATE INDEX idx_exam_documents_exam_type ON exam_documents(exam_type);

-- +migrate Down
DROP INDEX IF EXISTS idx_exam_documents_exam_type;
DROP INDEX IF EXISTS idx_exam_documents_status;
DROP INDEX IF EXISTS idx_exam_documents_patient_created_at;
DROP INDEX IF EXISTS idx_exam_documents_patient;
DROP TABLE IF EXISTS exam_documents;
