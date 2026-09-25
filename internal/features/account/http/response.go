// internal/features/account/http/response.go
package accounthttp

import (
	openapi "github.com/gabrielgcmr/sonnda/internal/api/openapi/generated"
	"github.com/gabrielgcmr/sonnda/internal/features/account"
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func userResponse(user *accountdomain.User) openapi.User {
	return openapi.User{
		Id:          user.ID,
		AuthIssuer:  user.AuthIssuer,
		AuthSubject: user.AuthSubject,
		Email:       user.Email,
		FullName:    user.FullName,
		AccountType: string(user.AccountType),
		BirthDate:   openapi_types.Date{Time: user.BirthDate},
		Cpf:         user.CPF,
		Phone:       user.Phone,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func patientsResponse(result *account.MyPatientsOutput) openapi.MyPatientsResponse {
	patients := make([]openapi.AccountPatientSummary, len(result.Patients))
	for i, patient := range result.Patients {
		patients[i] = openapi.AccountPatientSummary{
			Id:           patient.ID,
			FullName:     patient.FullName,
			AvatarUrl:    patient.AvatarURL,
			RelationType: patient.RelationType,
		}
	}
	return openapi.MyPatientsResponse{
		Patients: patients,
		Total:    result.Total,
		Limit:    result.Limit,
		Offset:   result.Offset,
	}
}
