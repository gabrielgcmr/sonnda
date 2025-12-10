package domain

type Gender string

const (
	GenderMale    Gender = "MALE"
	GenderFemale  Gender = "FEMALE"
	GenderOther   Gender = "OTHER"
	GenderUnknown Gender = "UNKNOWN"
)

type Race string

const (
	RaceWhite      Race = "WHITE"
	RaceBlack      Race = "BLACK"
	RaceAsian      Race = "ASIAN"
	RaceMixed      Race = "MIXED"
	RaceIndigenous Race = "INDIGENOUS"
	RaceUnknown    Race = "UNKNOWN"
)
