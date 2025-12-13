package domain

import (
	"time"

	"github.com/google/uuid"
)

type Patient struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	AppUserID *uuid.UUID `json:"app_user_id,omitempty" db:"app_user_id"`
	CPF       string     `json:"cpf" db:"cpf"`
	CNS       *string    `json:"cns,omitempty" db:"cns"`
	FullName  string     `json:"full_name" db:"full_name"`
	BirthDate time.Time  `json:"birth_date" db:"birth_date"`

	//Demograficos
	Gender Gender `json:"gender" db:"gender"`
	Race   Race   `json:"race" db:"race"`

	AvatarURL string    `json:"avatar_url" db:"avatar_url"`
	Phone     *string   `json:"phone,omitempty" db:"phone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Relacionamentos (Populed via JOIN ou queries separadas)
	MedicalRecords []MedicalRecord `json:"medical_records,omitempty"`
	Authorizations []Authorization `json:"authorizations,omitempty"`
}

// Mantendo a mesma estrutura do seu Kotlin
type MedicalRecord struct {
	ID          string            `json:"id" db:"id"`
	PatientID   string            `json:"patient_id" db:"patient_id"`
	CreatedBy   string            `json:"created_by" db:"created_by"` // User ID
	EntryType   MedicalRecordType `json:"entry_type" db:"entry_type"`
	Title       string            `json:"title" db:"title"`
	Description string            `json:"description" db:"description"`
	Date        time.Time         `json:"date" db:"date"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`

	// Dados específicos (JSONB no banco ou tabelas separadas)
	// Como saiu do GORM, você precisará gerenciar o carregamento destes dados manualmente
	PreventionData   *Prevention   `json:"prevention,omitempty"`
	ProblemData      *Problem      `json:"problem,omitempty"`
	LabReportData    *LabReport    `json:"lab_report,omitempty"`
	PhysicalExamData *PhysicalExam `json:"physical_exam,omitempty"`
}

type MedicalRecordType string

const (
	RecordTypePrevention   MedicalRecordType = "PREVENTION"
	RecordTypeProblem      MedicalRecordType = "PROBLEM"
	RecordTypeLabs         MedicalRecordType = "LABS"
	RecordTypeImageExam    MedicalRecordType = "IMAGE_EXAM"
	RecordTypePhysicalExam MedicalRecordType = "PHYSICAL_EXAM"
	RecordTypeNote         MedicalRecordType = "NOTE"
)

// Sub-estruturas
type Prevention struct {
	ID              string `json:"id" db:"id"`
	MedicalRecordID string `json:"medical_record_id" db:"medical_record_id"`
	Name            string `json:"name" db:"name"`
	Abbreviation    string `json:"abbreviation,omitempty" db:"abbreviation"`
	Value           string `json:"value,omitempty" db:"value"`
	ReferenceValue  string `json:"reference_value,omitempty" db:"reference_value"`
	Unit            string `json:"unit,omitempty" db:"unit"`
	Classification  string `json:"classification,omitempty" db:"classification"`
	Description     string `json:"description,omitempty" db:"description"`
	Other           string `json:"other,omitempty" db:"other"`
}

type Problem struct {
	ID              string `json:"id" db:"id"`
	MedicalRecordID string `json:"medical_record_id" db:"medical_record_id"`
	Name            string `json:"name" db:"name"`
	Abbreviation    string `json:"abbreviation,omitempty" db:"abbreviation"`
	BodySystem      string `json:"body_system,omitempty" db:"body_system"`
	Description     string `json:"description,omitempty" db:"description"`
	Other           string `json:"other,omitempty" db:"other"`
}

type PhysicalExam struct {
	ID              string `json:"id" db:"id"`
	MedicalRecordID string `json:"medical_record_id" db:"medical_record_id"`
	SystolicBP      string `json:"systolic_bp" db:"systolic_bp"`
	DiastolicBP     string `json:"diastolic_bp" db:"diastolic_bp"`
}

func NewPatient(
	appUserID *uuid.UUID,
	cpf string,
	cns *string,
	fullName string,
	birthDate time.Time,
	gender Gender,
	race Race,
	phone *string,
	avatarURL string,
) (*Patient, error) {
	// Validações de negócio
	normalizedCPF, err := NewCPF(cpf)
	if err != nil {
		return nil, err
	}

	if birthDate.After(time.Now()) {
		return nil, ErrInvalidBirthDate
	}
	// - etc.

	now := time.Now()

	return &Patient{
		ID:        uuid.Nil, // repo/banco preenche
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
	gender *Gender,
	race *Race,
	cns *string,
) {
	// Nome
	if fullName != nil && *fullName != "" {
		p.FullName = *fullName
	}

	// Telefone
	if phone != nil {
		p.Phone = phone
	}

	// Avatar
	if avatarURL != nil {
		p.AvatarURL = *avatarURL
	}

	// Gênero
	if gender != nil {
		p.Gender = *gender
	}

	// Raça
	if race != nil {
		p.Race = *race
	}

	// CNS
	if cns != nil {
		p.CNS = cns
	}

	p.UpdatedAt = time.Now()
}
