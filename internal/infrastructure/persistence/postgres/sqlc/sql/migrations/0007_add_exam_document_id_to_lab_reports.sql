-- +migrate Up
ALTER TABLE lab_reports
ADD COLUMN exam_document_id UUID REFERENCES exam_documents(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX idx_lab_reports_exam_document
ON lab_reports(exam_document_id)
WHERE exam_document_id IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_lab_reports_exam_document;

ALTER TABLE lab_reports
DROP COLUMN IF EXISTS exam_document_id;
