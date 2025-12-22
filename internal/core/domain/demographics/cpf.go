package demographics

import (
	"errors"
	"strings"
)

var ErrInvalidCPF = errors.New("CPF inválido")

type CPF string

// String retorna o CPF formatado
func (c CPF) String() string {
	return string(c)
}

// NewCPF é o construtor (Factory).
// Ele recebe uma string bruta, normaliza (limpa), valida e retorna o tipo do domínio.
func NewCPF(raw string) (CPF, error) {
	//1. Normalização
	s := strings.TrimSpace(raw)
	s = removeNonDigits(s)

	//2. Validação
	if len(s) != 11 || !isValidCPF(s) {
		return "", ErrInvalidCPF
	}
	return CPF(s), nil
}

func removeNonDigits(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func isValidCPF(s string) bool {
	if s == "00000000000" || s == "11111111111" || s == "22222222222" || s == "33333333333" ||
		s == "44444444444" || s == "55555555555" || s == "66666666666" || s == "77777777777" ||
		s == "88888888888" || s == "99999999999" {
		return false
	}

	var sum int
	for i := 0; i < 9; i++ {
		sum += int(s[i]-'0') * (10 - i)
	}

	remainder := sum % 11
	if remainder < 2 {
		if s[9] != '0' {
			return false
		}
	} else {
		if s[9] != byte(11-remainder+'0') {
			return false
		}
	}

	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(s[i]-'0') * (11 - i)
	}

	remainder = sum % 11
	if remainder < 2 {
		if s[10] != '0' {
			return false
		}
	} else {
		if s[10] != byte(11-remainder+'0') {
			return false
		}
	}

	return true
}
