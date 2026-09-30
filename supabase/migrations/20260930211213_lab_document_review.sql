-- supabase/migrations/20260930211213_lab_document_review.sql
ALTER TABLE public.exam_documents
 ADD COLUMN review_status TEXT CHECK (review_status IN ('pending', 'confirmed', 'deleting')),
 ADD COLUMN lab_report_id UUID REFERENCES public.lab_reports(id) ON DELETE RESTRICT,
 ADD COLUMN confirmed_by_user_id UUID REFERENCES public.users(id),
 ADD COLUMN confirmed_at TIMESTAMPTZ;
CREATE TABLE public.exam_document_extractions (
 document_id UUID PRIMARY KEY REFERENCES public.exam_documents(id) ON DELETE CASCADE,
 snapshot JSONB NOT NULL
);
ALTER TABLE public.exam_documents ADD CONSTRAINT exam_review_consistency CHECK (
 (review_status IS NULL AND lab_report_id IS NULL AND confirmed_by_user_id IS NULL AND confirmed_at IS NULL)
 OR (review_status IN ('pending','deleting') AND lab_report_id IS NULL AND confirmed_by_user_id IS NULL AND confirmed_at IS NULL)
 OR (review_status = 'confirmed' AND lab_report_id IS NOT NULL AND confirmed_by_user_id IS NOT NULL AND confirmed_at IS NOT NULL)
);
-- Snapshots are inserted once; confirmation never rewrites extracted values.
CREATE FUNCTION public.prevent_extraction_update() RETURNS trigger LANGUAGE plpgsql SET search_path = '' AS $$
BEGIN
 RAISE EXCEPTION 'Extraction snapshots are immutable';
END;
$$;
REVOKE ALL ON FUNCTION public.prevent_extraction_update() FROM PUBLIC, anon, authenticated;
CREATE TRIGGER immutable_extraction BEFORE UPDATE ON public.exam_document_extractions
 FOR EACH ROW EXECUTE FUNCTION public.prevent_extraction_update();
-- These resources are served exclusively by the authenticated Go API.
ALTER TABLE public.exam_document_extractions ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.exam_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.exam_document_texts ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON public.exam_document_extractions, public.exam_documents, public.exam_document_texts FROM anon, authenticated;
GRANT SELECT, INSERT, DELETE ON public.exam_document_extractions TO service_role;
