// internal/app/services/authorization/errors.go
package authorization

import "errors"

// Use seu padrão de erros (com code/status) se já existir.
// Aqui deixei simples e explícito.
var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not found")
)
