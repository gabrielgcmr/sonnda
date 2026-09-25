// internal/features/account/http/response.go
package accounthttp

import (
	accountdomain "github.com/gabrielgcmr/sonnda/internal/features/account/domain"
	openapi "github.com/gabrielgcmr/sonnda/internal/generated/openapi"
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
