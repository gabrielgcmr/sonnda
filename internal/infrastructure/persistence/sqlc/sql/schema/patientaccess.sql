-- PatientAccess: N:N relationship linking app users to patients with specific roles.
-- Mirrors the domain model: patient_id, user_id, role, timestamps.
CREATE TABLE patient_access (
    patient_id TEXT NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('caregiver', 'professional')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    PRIMARY KEY (patient_id, user_id)
);

-- Indexes for common lookups
CREATE INDEX idx_patient_access_user ON patient_access(user_id);
CREATE INDEX idx_patient_access_patient ON patient_access(patient_id);
