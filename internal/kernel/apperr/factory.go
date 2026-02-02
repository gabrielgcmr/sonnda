package apperr

// Fábricas de Erro (Factories)
// Validation retorna um erro deBadRequest (400)
func Validation(msg string, violations ...Violation) *AppError {
	return &AppError{
		Code:       VALIDATION_FAILED,
		Message:    msg,
		Violations: violations,
	}
}

// Internal retorna um erro de servidor (500) e oculta detalhes do cliente
func Internal(msg string, cause error) *AppError {
	return &AppError{
		Code:    INTERNAL_ERROR,
		Message: msg,
		Cause:   cause,
	}
}

// Conflict para recursos duplicados (409)
func Conflict(msg string) *AppError {
	return &AppError{
		Code:    RESOURCE_CONFLICT,
		Message: msg,
	}
}

func AlreadyExists(msg string) *AppError {
	return &AppError{
		Code:    RESOURCE_ALREADY_EXISTS,
		Message: msg,
	}
}

// Unauthorized para falhas de login (401)
func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:    AUTH_REQUIRED,
		Message: msg,
	}
}

// Forbidden para falhas de permissão (403)
func Forbidden(msg string) *AppError {
	return &AppError{
		Code:    ACCESS_DENIED,
		Message: msg,
	}
}
