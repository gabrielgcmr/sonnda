// internal/domain/entity/medicalrecord/antecedents/antecedents.go
package antecedents

import "time"

// Antecedents agrupa informações de histórico prévio do paciente.
// Mantém o modelo flexível (sem forçar codificações agora) para caber no MVP.
// Quando a persistência estiver definida, esses tipos podem virar tabelas/JSON.

type Antecedents struct {
	PastMedicalHistory PastMedicalHistory
	Lifestyle          LifestyleHabits
	FamilyHistory      FamilyHistory

	UpdatedAt *time.Time
}

// PastMedicalHistory cobre antecedentes pessoais (doenças prévias, cirurgias, etc.).

type PastMedicalHistory struct {
	Conditions       []HistoryItem // ex.: HAS, DM2, asma, etc.
	Surgeries        []HistoryItem
	Hospitalizations []HistoryItem
	Allergies        []HistoryItem
	Medications      []HistoryItem
	OtherNotes       *string
}

// LifestyleHabits cobre hábitos de vida.

type LifestyleHabits struct {
	Smoking          *string // ex.: "never", "former", "current" (ou texto livre)
	Alcohol          *string // ex.: "none", "social", "daily" (ou texto livre)
	PhysicalActivity *string
	Diet             *string
	Sleep            *string
	OtherNotes       *string
}

// FamilyHistory cobre antecedentes familiares relevantes.

type FamilyHistory struct {
	Items      []FamilyHistoryItem
	OtherNotes *string
}

type FamilyHistoryItem struct {
	Condition string  // ex.: "diabetes", "cancer"
	Relative  *string // ex.: "mother", "father" (ou texto livre)
	Notes     *string
}

type HistoryItem struct {
	Name      string
	SinceYear *int
	Notes     *string
}
