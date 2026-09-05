-- Alinha o nome generico usado pelo codigo com o schema remoto atual.
DO $$
BEGIN
  IF to_regclass('public.exam_reports') IS NOT NULL
     AND to_regclass('public.exam_document_texts') IS NULL THEN
    ALTER TABLE public.exam_reports RENAME TO exam_document_texts;
  END IF;

  IF to_regclass('public.exam_document_texts') IS NOT NULL
     AND EXISTS (
       SELECT 1
       FROM information_schema.columns
       WHERE table_schema = 'public'
         AND table_name = 'exam_document_texts'
         AND column_name = 'report_text'
     )
     AND NOT EXISTS (
       SELECT 1
       FROM information_schema.columns
       WHERE table_schema = 'public'
         AND table_name = 'exam_document_texts'
         AND column_name = 'text'
     ) THEN
    ALTER TABLE public.exam_document_texts RENAME COLUMN report_text TO text;
  END IF;
END $$;

-- Mantem os nomes de constraints coerentes depois do rename.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_reports_pkey'
      AND conrelid = 'public.exam_document_texts'::regclass
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_document_texts_pkey'
      AND conrelid = 'public.exam_document_texts'::regclass
  ) THEN
    ALTER TABLE public.exam_document_texts
      RENAME CONSTRAINT exam_reports_pkey TO exam_document_texts_pkey;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_reports_exam_document_id_key'
      AND conrelid = 'public.exam_document_texts'::regclass
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_document_texts_exam_document_id_key'
      AND conrelid = 'public.exam_document_texts'::regclass
  ) THEN
    ALTER TABLE public.exam_document_texts
      RENAME CONSTRAINT exam_reports_exam_document_id_key TO exam_document_texts_exam_document_id_key;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_reports_exam_document_id_fkey'
      AND conrelid = 'public.exam_document_texts'::regclass
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_document_texts_exam_document_id_fkey'
      AND conrelid = 'public.exam_document_texts'::regclass
  ) THEN
    ALTER TABLE public.exam_document_texts
      RENAME CONSTRAINT exam_reports_exam_document_id_fkey TO exam_document_texts_exam_document_id_fkey;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_reports_patient_id_fkey'
      AND conrelid = 'public.exam_document_texts'::regclass
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_document_texts_patient_id_fkey'
      AND conrelid = 'public.exam_document_texts'::regclass
  ) THEN
    ALTER TABLE public.exam_document_texts
      RENAME CONSTRAINT exam_reports_patient_id_fkey TO exam_document_texts_patient_id_fkey;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_reports_uploaded_by_user_id_fkey'
      AND conrelid = 'public.exam_document_texts'::regclass
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'exam_document_texts_uploaded_by_user_id_fkey'
      AND conrelid = 'public.exam_document_texts'::regclass
  ) THEN
    ALTER TABLE public.exam_document_texts
      RENAME CONSTRAINT exam_reports_uploaded_by_user_id_fkey TO exam_document_texts_uploaded_by_user_id_fkey;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_exam_reports_category'
      AND conrelid = 'public.exam_document_texts'::regclass
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_exam_document_texts_category'
      AND conrelid = 'public.exam_document_texts'::regclass
  ) THEN
    ALTER TABLE public.exam_document_texts
      RENAME CONSTRAINT chk_exam_reports_category TO chk_exam_document_texts_category;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_exam_reports_confidence'
      AND conrelid = 'public.exam_document_texts'::regclass
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_exam_document_texts_confidence'
      AND conrelid = 'public.exam_document_texts'::regclass
  ) THEN
    ALTER TABLE public.exam_document_texts
      RENAME CONSTRAINT chk_exam_reports_confidence TO chk_exam_document_texts_confidence;
  END IF;
END $$;

-- Renomeia indices legados quando existirem.
ALTER INDEX IF EXISTS public.idx_exam_reports_patient
  RENAME TO idx_exam_document_texts_patient;

ALTER INDEX IF EXISTS public.idx_exam_reports_patient_created_at
  RENAME TO idx_exam_document_texts_patient_created_at;

ALTER INDEX IF EXISTS public.idx_exam_reports_category
  RENAME TO idx_exam_document_texts_category;
