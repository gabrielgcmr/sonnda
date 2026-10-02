-- supabase/migrations/20261001234359_rename_lab_results_to_panels_observations.sql
-- Reconciles installations where lab_results was already renamed but its child table was not.
DO $$
BEGIN
    IF to_regclass('public.lab_result_items') IS NOT NULL
       AND to_regclass('public.observations') IS NULL THEN
        ALTER TABLE public.lab_result_items RENAME TO observations;
    END IF;

    IF to_regclass('public.lab_results') IS NOT NULL
       AND to_regclass('public.lab_panels') IS NULL THEN
        ALTER TABLE public.lab_results RENAME TO lab_panels;
    END IF;
END
$$;

DO $$
BEGIN
    IF to_regclass('public.observations') IS NOT NULL
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'observations' AND column_name = 'lab_result_id') THEN
        ALTER TABLE public.observations RENAME COLUMN lab_result_id TO lab_panel_id;
    END IF;

    IF to_regclass('public.observations') IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.observations') AND conname = 'observations_pkey')
       AND EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.observations') AND conname = 'lab_result_items_pkey') THEN
        ALTER TABLE public.observations RENAME CONSTRAINT lab_result_items_pkey TO observations_pkey;
    END IF;

    IF to_regclass('public.observations') IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.observations') AND conname = 'observations_lab_panel_id_fkey')
       AND EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.observations') AND conname = 'lab_result_items_lab_result_id_fkey') THEN
        ALTER TABLE public.observations RENAME CONSTRAINT lab_result_items_lab_result_id_fkey TO observations_lab_panel_id_fkey;
    END IF;

    IF to_regclass('public.lab_panels') IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.lab_panels') AND conname = 'lab_panels_pkey')
       AND EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.lab_panels') AND conname = 'lab_results_pkey') THEN
        ALTER TABLE public.lab_panels RENAME CONSTRAINT lab_results_pkey TO lab_panels_pkey;
    END IF;

    IF to_regclass('public.lab_panels') IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.lab_panels') AND conname = 'lab_panels_lab_report_id_fkey')
       AND EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.lab_panels') AND conname = 'lab_results_lab_report_id_fkey') THEN
        ALTER TABLE public.lab_panels RENAME CONSTRAINT lab_results_lab_report_id_fkey TO lab_panels_lab_report_id_fkey;
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS public.observations (
    id             UUID PRIMARY KEY,
    lab_panel_id   UUID NOT NULL REFERENCES public.lab_panels(id) ON DELETE CASCADE,
    parameter_name TEXT NOT NULL,
    result_value   TEXT,
    result_unit    TEXT,
    reference_text TEXT
);

DO $$
BEGIN
    IF to_regclass('public.idx_lab_result_items_result') IS NOT NULL
       AND to_regclass('public.idx_observations_panel') IS NULL THEN
        ALTER INDEX public.idx_lab_result_items_result RENAME TO idx_observations_panel;
    END IF;

    IF to_regclass('public.idx_lab_results_report') IS NOT NULL
       AND to_regclass('public.idx_lab_panels_report') IS NULL THEN
        ALTER INDEX public.idx_lab_results_report RENAME TO idx_lab_panels_report;
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_observations_panel ON public.observations(lab_panel_id);
CREATE INDEX IF NOT EXISTS idx_lab_panels_report ON public.lab_panels(lab_report_id);

ALTER TABLE public.observations ENABLE ROW LEVEL SECURITY;
