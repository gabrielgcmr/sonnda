package rbac

type Action string

const (
	//Recurso paciente
	ActionCreatePatient     Action = "patient:create"
	ActionSoftDeletePatient Action = "patient:soft_delete"
	ActionReadPatient       Action = "patient:read"
	ActionUpdatePatient     Action = "patient:update"
	//
	ActionRecordMeasurement Action = "measurement:record"
	ActionWriteClinicalNote Action = "clinical_note:write"
	// Exames laboratiriais do paciente
	ActionReadLabs   Action = "labs:read"
	ActionUploadLabs Action = "labs:upload"
	//Prescrições médicas do paciente
	ActionReadPrescriptions  Action = "prescriptions:read"
	ActionWritePrescriptions Action = "prescriptions:write"
)
