package user

import "strings"

type AccountType string

const (
	AccountTypeProfessional AccountType = "professional"
	AccountTypeBasicCare    AccountType = "basic_care"
	//AccountTypeAdmin        AccountType = "admin" // fora do MVP
)

func (at AccountType) Normalize() AccountType {
	return AccountType(strings.ToLower(strings.TrimSpace(string(at))))
}

func (at AccountType) IsValid() bool {
	switch at {
	case AccountTypeProfessional, AccountTypeBasicCare:
		return true
	default:
		return false
	}
}

func ParseAccountType(raw string) (AccountType, error) {
	at := AccountType(strings.ToLower(strings.TrimSpace(raw)))
	if !at.IsValid() {
		return "", ErrInvalidAccountType
	}
	return at, nil

}
