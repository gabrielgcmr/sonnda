package patient

import (
	"time"

	"sonnda-api/internal/core/domain/demographics"
)

type Patient struct {
	ID        string
	AppUserID *string
	CPF       string
	CNS       *string
	FullName  string
	BirthDate time.Time

	Gender demographics.Gender
	Race   demographics.Race

	AvatarURL string
	Phone     *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPatient(
	appUserID *string,
	cpf string,
	cns *string,
	fullName string,
	birthDate time.Time,
	gender demographics.Gender,
	race demographics.Race,
	phone *string,
	avatarURL string,
) (*Patient, error) {
	normalizedCPF, err := demographics.NewCPF(cpf)
	if err != nil {
		return nil, err
	}

	if birthDate.After(time.Now()) {
		return nil, ErrInvalidBirthDate
	}

	now := time.Now()

	return &Patient{
		ID:        "",
		CPF:       normalizedCPF.String(),
		CNS:       cns,
		FullName:  fullName,
		BirthDate: birthDate,
		AppUserID: appUserID,
		Gender:    gender,
		Race:      race,
		AvatarURL: avatarURL,
		Phone:     phone,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *Patient) ApplyUpdate(
	fullName *string,
	phone *string,
	avatarURL *string,
	gender *demographics.Gender,
	race *demographics.Race,
	cns *string,
) {
	if fullName != nil && *fullName != "" {
		p.FullName = *fullName
	}

	if phone != nil {
		p.Phone = phone
	}

	if avatarURL != nil {
		p.AvatarURL = *avatarURL
	}

	if gender != nil {
		p.Gender = *gender
	}

	if race != nil {
		p.Race = *race
	}

	if cns != nil {
		p.CNS = cns
	}

	p.UpdatedAt = time.Now()
}
