// internal/app/apperr/helpers.go
package apperr

import "errors"

// New cria um erro da aplicação sem causa interna.
func New(kind Kind, code, message string) *Error {
	return &Error{
		Kind:    kind,
		Code:    code,
		Message: message,
	}
}

// Wrap cria um erro da aplicação encapsulando outro erro.
func Wrap(kind Kind, code string, err error) *Error {
	return &Error{
		Kind: kind,
		Code: code,
		Err:  err,
	}
}

// IsKind verifica se o erro (ou algum erro na cadeia) tem o Kind informado.
func IsKind(err error, kind Kind) bool {
	var ae *Error
	if errors.As(err, &ae) {
		return ae.Kind == kind
	}
	return false
}

// As retorna o *apperr.Error, se existir na cadeia.
func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
