package patient

type CreatePatientRequest struct {
	CPF       string  `json:"cpf" binding:"required"`
	FullName  string  `json:"full_name" binding:"required"`
	BirthDate string  `json:"birth_date" binding:"required,datetime=2006-01-02"`
	Gender    string  `json:"gender" binding:"required,oneof=MALE FEMALE OTHER UNKNOWN"`
	Race      string  `json:"race" binding:"required,oneof=WHITE BLACK ASIAN MIXED INDIGENOUS UNKNOWN"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL string  `json:"avatar_url"`
}

type updatePatientRequest struct {
	// Permite PATCH-like via PUT (campos opcionais).
	FullName  *string `json:"full_name,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"` // yyyy-mm-dd
	Gender    *string `json:"gender,omitempty"`
	Race      *string `json:"race,omitempty"`

	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}
