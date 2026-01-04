package apperr

type Violation struct {
	Field  string //ex: "professional.email"
	Reason string //ex: "required", "invalid email format"
}
