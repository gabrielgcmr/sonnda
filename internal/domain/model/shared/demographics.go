package shared

import (
	"fmt"
	"strings"
)

type Gender string

const (
	GenderMale    Gender = "MALE"
	GenderFemale  Gender = "FEMALE"
	GenderOther   Gender = "OTHER"
	GenderUnknown Gender = "UNKNOWN"
)

func ParseGender(input string) (Gender, error) {
	value := strings.ToUpper(strings.TrimSpace(input))

	switch value {
	case string(GenderMale):
		return GenderMale, nil
	case string(GenderFemale):
		return GenderFemale, nil
	case string(GenderOther):
		return GenderOther, nil
	case string(GenderUnknown):
		return GenderUnknown, nil
	default:
		return "", fmt.Errorf("invalid gender: %s", input)
	}
}

type Race string

const (
	RaceWhite      Race = "WHITE"
	RaceBlack      Race = "BLACK"
	RaceAsian      Race = "ASIAN"
	RaceMixed      Race = "MIXED"
	RaceIndigenous Race = "INDIGENOUS"
	RaceUnknown    Race = "UNKNOWN"
)

func ParseRace(input string) (Race, error) {
	value := strings.ToUpper(strings.TrimSpace(input))

	switch value {
	case string(RaceWhite):
		return RaceWhite, nil
	case string(RaceBlack):
		return RaceBlack, nil
	case string(RaceAsian):
		return RaceAsian, nil
	case string(RaceMixed):
		return RaceMixed, nil
	case string(RaceIndigenous):
		return RaceIndigenous, nil
	case string(RaceUnknown):
		return RaceUnknown, nil
	default:
		return "", fmt.Errorf("invalid race: %s", input)
	}
}

func CleanDigits(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
