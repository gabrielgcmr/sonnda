package rbac

type Action string

const (
	//Recurso paciente (Foque aqui no MVP)
	ActionCreatePatient      Action = "patient:create"
	ActionUpdatePatient      Action = "patient:update"
	ActionSoftDeletePatient  Action = "patient:soft_delete"
	ActionViewMinimalPatient Action = "patient:view_minimal"
	ActionViewFullPatient    Action = "patient:view_full"

	// Dados cadastrais do paciente
	ActionReadPatientDemographics   Action = "patient:read"
	ActionWritePatientDemographics  Action = "patient:write"
	ActionUpdatePatientDemographics Action = "patient:update"
	// Exames laboratiriais do paciente
	ActionReadLabs   Action = "labs:read"
	ActionUploadLabs Action = "labs:upload"
	//Prescrições médicas do paciente
	ActionReadPrescriptions  Action = "prescriptions:read"
	ActionWritePrescriptions Action = "prescriptions:write"
)
