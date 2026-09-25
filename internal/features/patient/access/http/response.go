// internal/features/patient/access/http/response.go
package accesshttp

import (
	openapi "github.com/gabrielgcmr/sonnda/internal/api/openapi/generated"
	patientaccess "github.com/gabrielgcmr/sonnda/internal/features/patient/access"
)

func listPatientsResponse(result *patientaccess.ListPatientsOutput) openapi.AccessiblePatientsResponse {
	patients := make([]openapi.AccessiblePatientSummary, len(result.Patients))
	for i, patient := range result.Patients {
		patients[i] = openapi.AccessiblePatientSummary{
			Id:        patient.ID,
			FullName:  patient.FullName,
			AvatarUrl: patient.AvatarURL,
		}
	}
	return openapi.AccessiblePatientsResponse{
		Patients: patients,
		Total:    result.Total,
		Limit:    result.Limit,
		Offset:   result.Offset,
	}
}
