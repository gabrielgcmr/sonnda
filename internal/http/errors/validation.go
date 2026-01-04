package errors

import (
	"errors"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"

	"sonnda-api/internal/app/apperr"
)

// ValidationErrorsToViolations converte erros do validator em violações agnósticas.
func ValidationErrorsToViolations(err error) []apperr.Violation {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return nil
	}

	violations := make([]apperr.Violation, 0, len(verrs))

	for _, fe := range verrs {
		violations = append(violations, apperr.Violation{
			Field:  fieldPath(fe),
			Reason: reasonFromTag(fe.Tag()),
		})
	}

	return violations
}

func fieldPath(fe validator.FieldError) string {
	// Usa o nome do campo no JSON, não o nome do struct
	ns := fe.Namespace() // ex: RegisterRequest.Professional.RegistrationNumber

	parts := strings.Split(ns, ".")
	if len(parts) <= 1 {
		return fe.Field()
	}

	// remove o nome do struct raiz
	parts = parts[1:]

	for i := range parts {
		parts[i] = toSnakeCase(parts[i])
	}

	return strings.Join(parts, ".")
}

func reasonFromTag(tag string) string {
	switch tag {
	case "required", "required_if":
		return "required"
	case "email":
		return "invalid_email"
	case "oneof":
		return "invalid_value"
	case "datetime":
		return "invalid_date"
	case "min":
		return "min"
	case "max":
		return "max"
	case "len":
		return "invalid_length"
	case "uuid":
		return "invalid_uuid"
	default:
		return tag
	}
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
