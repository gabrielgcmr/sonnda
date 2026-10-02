-- supabase/migrations/20261001234449_reconcile_observation_table.sql
-- Reconciles the singular table left by an earlier partial rename.
DO $$
DECLARE
    empty_target BOOLEAN;
BEGIN
    IF to_regclass('public.observation') IS NULL THEN
        RETURN;
    END IF;

    IF to_regclass('public.observations') IS NOT NULL THEN
        EXECUTE 'SELECT count(*) = 0 FROM public.observations' INTO empty_target;
        IF NOT empty_target THEN
            RAISE EXCEPTION 'Cannot reconcile observation tables: observations already contains data';
        END IF;
        DROP TABLE public.observations;
    END IF;

    ALTER TABLE public.observation RENAME TO observations;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'observations' AND column_name = 'lab_result_id') THEN
        ALTER TABLE public.observations RENAME COLUMN lab_result_id TO lab_panel_id;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.observations') AND conname = 'lab_result_items_pkey') THEN
        ALTER TABLE public.observations RENAME CONSTRAINT lab_result_items_pkey TO observations_pkey;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = to_regclass('public.observations') AND conname = 'lab_result_items_lab_result_id_fkey') THEN
        ALTER TABLE public.observations RENAME CONSTRAINT lab_result_items_lab_result_id_fkey TO observations_lab_panel_id_fkey;
    END IF;
    IF to_regclass('public.idx_observations_panel') IS NULL AND to_regclass('public.idx_lab_result_items_result') IS NOT NULL THEN
        ALTER INDEX public.idx_lab_result_items_result RENAME TO idx_observations_panel;
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_observations_panel ON public.observations(lab_panel_id);
ALTER TABLE public.observations ENABLE ROW LEVEL SECURITY;
