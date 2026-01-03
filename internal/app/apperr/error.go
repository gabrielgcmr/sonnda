// internal/app/apperr/error.go
package apperr

type AppError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	return e.Message
}

// Permite errors.Is /errors.As funcionar corretamente
func (e *AppError) Unwrap() error {
	return e.Cause
}
