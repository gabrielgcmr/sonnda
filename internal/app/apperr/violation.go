package apperr

type Violation struct {
	Field  string `json:"field"`  // ex: "professional.email"
	Reason string `json:"reason"` // ex: "required", "invalid_email"
}
