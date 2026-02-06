// internal/kernel/apperr/error.go
package apperr

type Violation struct {
	Field  string `json:"field"`  // ex: "professional.email"
	Reason string `json:"reason"` // ex: "required", "invalid_email"
}

type AppError struct {
	Kind       ErrorKind
	Message    string
	Cause      error
	Violations []Violation
}

func (e *AppError) Error() string {
	return e.Message
}

// Permite errors.Is /errors.As funcionar corretamente
func (e *AppError) Unwrap() error {
	return e.Cause
}
