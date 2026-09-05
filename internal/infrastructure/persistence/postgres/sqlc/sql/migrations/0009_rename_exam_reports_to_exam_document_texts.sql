-- +migrate Up
ALTER TABLE exam_reports RENAME TO exam_document_texts;
ALTER TABLE exam_document_texts RENAME COLUMN report_text TO text;

ALTER TABLE exam_document_texts
    RENAME CONSTRAINT chk_exam_reports_category TO chk_exam_document_texts_category;

ALTER TABLE exam_document_texts
    RENAME CONSTRAINT chk_exam_reports_confidence TO chk_exam_document_texts_confidence;

ALTER INDEX idx_exam_reports_patient RENAME TO idx_exam_document_texts_patient;
ALTER INDEX idx_exam_reports_patient_created_at RENAME TO idx_exam_document_texts_patient_created_at;
ALTER INDEX idx_exam_reports_category RENAME TO idx_exam_document_texts_category;

-- +migrate Down
ALTER INDEX idx_exam_document_texts_category RENAME TO idx_exam_reports_category;
ALTER INDEX idx_exam_document_texts_patient_created_at RENAME TO idx_exam_reports_patient_created_at;
ALTER INDEX idx_exam_document_texts_patient RENAME TO idx_exam_reports_patient;

ALTER TABLE exam_document_texts
    RENAME CONSTRAINT chk_exam_document_texts_confidence TO chk_exam_reports_confidence;

ALTER TABLE exam_document_texts
    RENAME CONSTRAINT chk_exam_document_texts_category TO chk_exam_reports_category;

ALTER TABLE exam_document_texts RENAME COLUMN text TO report_text;
ALTER TABLE exam_document_texts RENAME TO exam_reports;
