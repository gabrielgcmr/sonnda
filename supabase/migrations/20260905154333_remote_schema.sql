SET local check_function_bodies = off;

CREATE EXTENSION "pg_trgm" SCHEMA "public";

CREATE TABLE "public"."exam_documents" (
  "id"                  uuid                     NOT NULL,
  "patient_id"          uuid                     NOT NULL,
  "uploaded_by_user_id" uuid                     NOT NULL,
  "storage_uri"         text                     NOT NULL,
  "original_filename"   text                     NOT NULL,
  "mime_type"           text                     NOT NULL,
  "status"              text                     NOT NULL,
  "exam_type"           text,
  "extraction_method"   text,
  "confidence"          double precision,
  "extracted_text"      text,
  "error_message"       text,
  "created_at"          timestamp with time zone NOT NULL DEFAULT now(),
  "updated_at"          timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT "chk_exam_documents_confidence" CHECK (((confidence IS NULL) OR ((confidence >= (0)::double precision) AND (confidence <= (1)::double precision)))),
  CONSTRAINT "chk_exam_documents_exam_type" CHECK (((exam_type IS NULL) OR (exam_type = ANY (ARRAY['laboratory'::text, 'imaging'::text, 'unknown'::text])))),
  CONSTRAINT "chk_exam_documents_status" CHECK ((status = ANY (ARRAY['uploaded'::text, 'processing'::text, 'processed'::text, 'failed'::text, 'needs_review'::text]))),
  CONSTRAINT "exam_documents_pkey" PRIMARY KEY (id)
);

CREATE TABLE "public"."exam_reports" (
  "id"                  uuid                     NOT NULL,
  "exam_document_id"    uuid,
  "patient_id"          uuid                     NOT NULL,
  "uploaded_by_user_id" uuid                     NOT NULL,
  "category"            text                     NOT NULL,
  "title"               text,
  "modality"            text,
  "body_site"           text,
  "performed_at"        timestamp with time zone,
  "facility_name"       text,
  "interpreting_doctor" text,
  "report_text"         text                     NOT NULL,
  "conclusion"          text,
  "extraction_method"   text,
  "confidence"          double precision,
  "created_at"          timestamp with time zone NOT NULL DEFAULT now(),
  "updated_at"          timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT "exam_reports_exam_document_id_key" UNIQUE (exam_document_id),
  CONSTRAINT "exam_reports_pkey" PRIMARY KEY (id)
);

CREATE TABLE "public"."lab_reports" (
  "id"                  uuid                     NOT NULL,
  "patient_id"          uuid                     NOT NULL,
  "uploaded_by_user_id" uuid                     NOT NULL,
  "patient_name"        text,
  "patient_dob"         timestamp with time zone,
  "lab_name"            text,
  "lab_phone"           text,
  "insurance_provider"  text,
  "requesting_doctor"   text,
  "technical_manager"   text,
  "report_date"         timestamp with time zone,
  "raw_text"            text,
  "fingerprint"         text,
  "created_at"          timestamp with time zone NOT NULL DEFAULT now(),
  "updated_at"          timestamp with time zone NOT NULL DEFAULT now(),
  "exam_document_id"    uuid,
  CONSTRAINT "lab_reports_pkey" PRIMARY KEY (id)
);

ALTER TABLE "public"."lab_reports"
  ENABLE ROW LEVEL SECURITY;

CREATE TABLE "public"."lab_result_items" (
  "id"             uuid NOT NULL,
  "lab_result_id"  uuid NOT NULL,
  "parameter_name" text NOT NULL,
  "result_value"   text,
  "result_unit"    text,
  "reference_text" text,
  CONSTRAINT "lab_result_items_pkey" PRIMARY KEY (id)
);

ALTER TABLE "public"."lab_result_items"
  ENABLE ROW LEVEL SECURITY;

CREATE TABLE "public"."lab_results" (
  "id"            uuid                     NOT NULL,
  "lab_report_id" uuid                     NOT NULL,
  "test_name"     text                     NOT NULL,
  "material"      text,
  "method"        text,
  "collected_at"  timestamp with time zone,
  "release_at"    timestamp with time zone,
  CONSTRAINT "lab_results_pkey" PRIMARY KEY (id)
);

ALTER TABLE "public"."lab_results"
  ENABLE ROW LEVEL SECURITY;

CREATE TABLE "public"."patient_access" (
  "patient_id"    uuid                     NOT NULL,
  "grantee_id"    uuid                     NOT NULL,
  "relation_type" text                     NOT NULL,
  "created_at"    timestamp with time zone NOT NULL DEFAULT now(),
  "revoked_at"    timestamp with time zone,
  "granted_by"    uuid,
  CONSTRAINT "patient_access_pkey" PRIMARY KEY (patient_id, grantee_id),
  CONSTRAINT "patient_access_relation_type_check" CHECK ((relation_type = ANY (ARRAY['caregiver'::text, 'family'::text, 'professional'::text, 'self'::text])))
);

ALTER TABLE "public"."patient_access"
  ENABLE ROW LEVEL SECURITY;

CREATE TABLE "public"."patients" (
  "id"                 uuid                     NOT NULL,
  "owner_user_id"      uuid,
  "cpf"                text                     NOT NULL,
  "cns"                text,
  "full_name"          text                     NOT NULL,
  "birth_date"         date                     NOT NULL,
  "gender"             text                     NOT NULL,
  "race"               text                     NOT NULL,
  "phone"              text,
  "avatar_url"         text,
  "created_at"         timestamp with time zone NOT NULL DEFAULT now(),
  "updated_at"         timestamp with time zone NOT NULL DEFAULT now(),
  "deleted_at"         timestamp with time zone,
  "created_by_user_id" uuid,
  CONSTRAINT "chk_patients_gender" CHECK ((gender = ANY (ARRAY['MALE'::text, 'FEMALE'::text, 'OTHER'::text, 'UNKNOWN'::text]))),
  CONSTRAINT "chk_patients_race" CHECK ((race = ANY (ARRAY['WHITE'::text, 'BLACK'::text, 'ASIAN'::text, 'MIXED'::text, 'INDIGENOUS'::text, 'UNKNOWN'::text]))),
  CONSTRAINT "patients_cpf_key" UNIQUE (cpf),
  CONSTRAINT "patients_pkey" PRIMARY KEY (id),
  CONSTRAINT "patients_user_id_key" UNIQUE (owner_user_id)
);

ALTER TABLE "public"."patients"
  ENABLE ROW LEVEL SECURITY;

CREATE TABLE "public"."professionals" (
  "user_id"             uuid                     NOT NULL,
  "registration_number" text                     NOT NULL,
  "registration_issuer" text                     NOT NULL,
  "registration_state"  text,
  "status"              text                     NOT NULL,
  "verified_at"         timestamp with time zone,
  "created_at"          timestamp with time zone NOT NULL DEFAULT now(),
  "updated_at"          timestamp with time zone NOT NULL DEFAULT now(),
  "deleted_at"          timestamp with time zone,
  "kind"                text                     NOT NULL,
  CONSTRAINT "professionals_pkey" PRIMARY KEY (user_id),
  CONSTRAINT "professionals_status_check" CHECK ((status = ANY (ARRAY['pending'::text, 'verified'::text, 'rejected'::text])))
);

ALTER TABLE "public"."professionals"
  ENABLE ROW LEVEL SECURITY;

CREATE TABLE "public"."users" (
  "id"           uuid                     NOT NULL,
  "auth_issuer"  text                     NOT NULL,
  "auth_subject" text                     NOT NULL,
  "email"        text                     NOT NULL,
  "full_name"    text                     NOT NULL,
  "birth_date"   date                     NOT NULL,
  "cpf"          text                     NOT NULL,
  "phone"        text                     NOT NULL,
  "created_at"   timestamp with time zone NOT NULL DEFAULT now(),
  "updated_at"   timestamp with time zone NOT NULL DEFAULT now(),
  "deleted_at"   timestamp with time zone,
  "account_type" text                     NOT NULL DEFAULT 'basic_care'::text,
  CONSTRAINT "users_cpf_key" UNIQUE (cpf),
  CONSTRAINT "users_email_key" UNIQUE (email),
  CONSTRAINT "users_pkey" PRIMARY KEY (id)
);

ALTER TABLE "public"."users"
  ENABLE ROW LEVEL SECURITY;

CREATE TYPE "public"."gender_enum" AS ENUM (
  'MALE',
  'FEMALE',
  'OTHER',
  'UNKNOWN'
);

CREATE TYPE "public"."race_enum" AS ENUM (
  'WHITE',
  'BLACK',
  'ASIAN',
  'MIXED',
  'INDIGENOUS',
  'UNKNOWN'
);

CREATE OR REPLACE FUNCTION public.create_patient_creator_access()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path TO 'public'
  AS $function$
begin
  insert into public.patient_access (
    patient_id,
    grantee_id,
    relation_type,
    granted_by
  )
  values (
    new.id,
    new.created_by_user_id,
    case
      when new.owner_user_id = new.created_by_user_id then 'self'
      else 'creator'
    end,
    new.created_by_user_id
  );

  return new;
end;
$function$;

CREATE OR REPLACE FUNCTION public.current_app_user_id()
  RETURNS uuid
  LANGUAGE sql
  STABLE
  SECURITY DEFINER
  SET search_path TO 'public'
  AS $function$select u.id
  from public.users u
  where u.auth_subject = auth.uid()::text
  limit 1$function$;

ALTER TABLE "public"."exam_reports"
  ADD CONSTRAINT "exam_reports_exam_document_id_fkey" FOREIGN KEY (exam_document_id) REFERENCES public.exam_documents(id) ON DELETE SET NULL;

ALTER TABLE "public"."lab_reports"
  ADD CONSTRAINT "lab_reports_exam_document_id_fkey" FOREIGN KEY (exam_document_id) REFERENCES public.exam_documents(id) ON DELETE SET NULL;

ALTER TABLE "public"."lab_results"
  ADD CONSTRAINT "lab_results_lab_report_id_fkey" FOREIGN KEY (lab_report_id) REFERENCES public.lab_reports(id) ON DELETE CASCADE;

ALTER TABLE "public"."lab_result_items"
  ADD CONSTRAINT "lab_result_items_lab_result_id_fkey" FOREIGN KEY (lab_result_id) REFERENCES public.lab_results(id) ON DELETE CASCADE;

ALTER TABLE "public"."exam_documents"
  ADD CONSTRAINT "exam_documents_patient_id_fkey" FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;

ALTER TABLE "public"."exam_reports"
  ADD CONSTRAINT "exam_reports_patient_id_fkey" FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;

ALTER TABLE "public"."lab_reports"
  ADD CONSTRAINT "lab_reports_patient_id_fkey" FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;

ALTER TABLE "public"."patient_access"
  ADD CONSTRAINT "patient_access_patient_id_fkey" FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;

ALTER TABLE "public"."exam_documents"
  ADD CONSTRAINT "exam_documents_uploaded_by_user_id_fkey" FOREIGN KEY (uploaded_by_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE "public"."exam_reports"
  ADD CONSTRAINT "exam_reports_uploaded_by_user_id_fkey" FOREIGN KEY (uploaded_by_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE "public"."lab_reports"
  ADD CONSTRAINT "lab_reports_uploaded_by_user_id_fkey" FOREIGN KEY (uploaded_by_user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE "public"."patient_access"
  ADD CONSTRAINT "patient_access_granted_by_fkey" FOREIGN KEY (granted_by) REFERENCES public.users(id);

ALTER TABLE "public"."patient_access"
  ADD CONSTRAINT "patient_access_grantee_id_fkey" FOREIGN KEY (grantee_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE "public"."patients"
  ADD CONSTRAINT "patients_created_by_user_id_fkey" FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

ALTER TABLE "public"."patients"
  ADD CONSTRAINT "patients_owner_user_id_fkey" FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE "public"."professionals"
  ADD CONSTRAINT "professionals_user_id_fkey" FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

CREATE INDEX idx_exam_documents_exam_type ON public.exam_documents USING btree (exam_type);

CREATE INDEX idx_exam_documents_patient_created_at ON public.exam_documents USING btree (patient_id, created_at DESC);

CREATE INDEX idx_exam_documents_patient ON public.exam_documents USING btree (patient_id);

CREATE INDEX idx_exam_documents_status ON public.exam_documents USING btree (status);

CREATE UNIQUE INDEX idx_lab_reports_exam_document ON public.lab_reports USING btree (exam_document_id)
  WHERE (exam_document_id IS NOT NULL);

CREATE UNIQUE INDEX idx_lab_reports_fingerprint ON public.lab_reports USING btree (fingerprint)
  WHERE (fingerprint IS NOT NULL);

CREATE INDEX idx_lab_reports_patient ON public.lab_reports USING btree (patient_id);

CREATE INDEX idx_lab_reports_report_date ON public.lab_reports USING btree (report_date);

CREATE INDEX idx_lab_result_items_result ON public.lab_result_items USING btree (lab_result_id);

CREATE INDEX idx_lab_results_report ON public.lab_results USING btree (lab_report_id);

CREATE INDEX idx_patient_access_active ON public.patient_access USING btree (grantee_id, patient_id)
  WHERE (revoked_at IS NULL);

CREATE INDEX idx_patient_access_patient ON public.patient_access USING btree (patient_id);

CREATE INDEX idx_patient_access_user ON public.patient_access USING btree (grantee_id);

CREATE INDEX idx_patients_cns ON public.patients USING btree (cns);

CREATE INDEX idx_patients_cpf ON public.patients USING btree (cpf);

CREATE INDEX idx_patients_full_name_trgm ON public.patients USING gin (full_name public.gin_trgm_ops);

CREATE INDEX idx_professionals_deleted_at ON public.professionals USING btree (deleted_at);

CREATE INDEX idx_professionals_reg_number ON public.professionals USING btree (registration_number);

CREATE INDEX idx_professionals_status ON public.professionals USING btree (status);

CREATE UNIQUE INDEX idx_users_auth_identity ON public.users USING btree (auth_issuer, auth_subject);

CREATE INDEX idx_users_cpf ON public.users USING btree (cpf);

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);

CREATE INDEX idx_users_email ON public.users USING btree (email);

CREATE UNIQUE INDEX ux_patients_cpf_active ON public.patients USING btree (cpf)
  WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX ux_patients_user_id_active ON public.patients USING btree (owner_user_id)
  WHERE (deleted_at IS NULL);

CREATE POLICY "users can create patients" ON "public"."patients"
  FOR INSERT
  TO "authenticated"
  WITH CHECK (((created_by_user_id = public.current_app_user_id()) AND ((owner_user_id IS NULL) OR (owner_user_id = public.current_app_user_id()))));

CREATE POLICY "users can view linked patients" ON "public"."patients"
  FOR SELECT
  TO "authenticated"
  USING ((EXISTS ( SELECT 1
   FROM public.patient_access pa
  WHERE ((pa.patient_id = patients.id) AND (pa.grantee_id = public.current_app_user_id()) AND (pa.revoked_at IS NULL)))));

CREATE POLICY "Users can CRUD own profile" ON "public"."users"
  FOR ALL
  TO "authenticated"
  USING ((auth_subject = (auth.uid())::text))
  WITH CHECK ((auth_subject = (auth.uid())::text));

COMMENT ON EXTENSION "pg_trgm" IS 'text similarity measurement and index searching based on trigrams';

COMMENT ON TYPE "public"."gender_enum" IS 'generos';

COMMENT ON TYPE "public"."race_enum" IS 'Raças disponíveis';

GRANT EXECUTE ON FUNCTION "public"."create_patient_creator_access"() TO PUBLIC, "anon", "authenticated", "postgres", "service_role";

GRANT EXECUTE ON FUNCTION "public"."current_app_user_id"() TO PUBLIC, "anon", "authenticated", "postgres", "service_role";

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE "public"."exam_documents" TO "anon", "authenticated", "postgres", "service_role";

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE "public"."exam_reports" TO "anon", "authenticated", "postgres", "service_role";

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE "public"."lab_reports" TO "anon", "authenticated", "postgres", "service_role";

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE "public"."lab_result_items" TO "anon", "authenticated", "postgres", "service_role";

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE "public"."lab_results" TO "anon", "authenticated", "postgres", "service_role";

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE "public"."patient_access" TO "anon", "authenticated", "postgres", "service_role";

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE "public"."patients" TO "anon", "authenticated", "postgres", "service_role";

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE "public"."professionals" TO "anon", "authenticated", "postgres", "service_role";

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE "public"."users" TO "anon", "authenticated", "postgres", "service_role";

GRANT USAGE ON TYPE "public"."gender_enum" TO "postgres";

GRANT USAGE ON TYPE "public"."race_enum" TO "postgres";

