package domain

import "errors"

// auth
var (
	ErrInvalidAuthProvider = errors.New("o provedor de autenticação é obrigatório")
	ErrInvalidAuthSubject  = errors.New("o identificador de autenticação (subject) é obrigatório")
	ErrInvalidEmail        = errors.New("o e-mail é obrigatório ou inválido")
	ErrInvalidRole         = errors.New("o perfil de usuário (role) fornecido é inválido")
	ErrEmailAlreadyExists  = errors.New("e-mail já cadastrado")
)

// authorization
var (
	ErrForbidden = errors.New("ação proibida: usuário não autorizado")
)

// patient
var (
	ErrPatientNotFound  = errors.New("paciente não encontrado")
	ErrCPFAlreadyExists = errors.New("CPF já cadastrado")

	ErrInvalidBirthDate = errors.New("data de nascimento inválida")
	ErrPatientTooYoung  = errors.New("paciente deve ter pelo menos 18 anos")
)

// labs
var (
	// ... erros existentes
	ErrLabReportNotFound      = errors.New("laudo não encontrado")
	ErrInvalidDocument        = errors.New("documento inválido")
	ErrDocumentProcessing     = errors.New("erro ao processar documento")
	ErrInvalidDateFormat      = errors.New("formato de data inválido")
	ErrInvalidInput           = errors.New("formato de input inválido")
	ErrMissingIdentifiers     = errors.New("identificadores obrigatórios ausentes")
	ErrLabReportAlreadyExists = errors.New("lab report already exists")
)
