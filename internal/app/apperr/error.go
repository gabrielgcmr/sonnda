// internal/app/apperr/error.go
package apperr

// Error é o erro padrão da aplicação.
//
// Ele carrega:
// - Kind: significado semântico
// - Code: identificador estável para clientes (ex.: "patient_not_found")
// - Message: descrição humana (opcional)
// - Err: causa original (wrap)
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *Error) Unwrap() error {
	return e.Err
}
