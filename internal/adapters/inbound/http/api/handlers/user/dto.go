package user

type RegisterRequest struct {
	FullName     string                   `json:"full_name" binding:"required"`
	BirthDate    string                   `json:"birth_date" binding:"required,datetime=2006-01-02"` // Gin já valida data!
	CPF          string                   `json:"cpf" binding:"required"`
	Phone        string                   `json:"phone" binding:"required"`
	AccountType  string                   `json:"account_type" binding:"required,oneof=basic_care professional"`
	Professional *ProfessionalRequestData `json:"professional" binding:"required_if=AccountType professional"` // Magia do Gin
}
type ProfessionalRequestData struct {
	Kind               string  `json:"kind" binding:"required"`
	RegistrationNumber string  `json:"registration_number" binding:"required"`
	RegistrationIssuer string  `json:"registration_issuer" binding:"required"`
	RegistrationState  *string `json:"registration_state,omitempty"`
}
type UpdateUserRequest struct {
	FullName  *string `json:"full_name,omitempty"`
	BirthDate *string `json:"birth_date,omitempty" binding:"required,datetime=2006-01-02"`
	CPF       *string `json:"cpf,omitempty"`
	Phone     *string `json:"phone,omitempty"`
}
