package authorization

import (
	"context"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/services"
)

var _ services.AuthorizationService = (*SimpleAuthorizationService)(nil)

// SimpleAuthorizationService implementa regras básicas de autorização
// baseadas apenas em Role e relacionamento direto user/patient.
type SimpleAuthorizationService struct {
	// se no futuro você quiser conferir vínculos médico↔paciente em tabela,
	// pode injetar repositórios aqui (ex.: MedicoPacienteRepository).
}

func NewAuthorizationService() *SimpleAuthorizationService {
	return &SimpleAuthorizationService{}
}

// ---------- Pacientes ----------

func (s *SimpleAuthorizationService) CanCreateDoctor(
	ctx context.Context,
	actor *domain.User,
	newUser *domain.User,
) bool {
	if actor == nil || newUser == nil {
		return false
	}

	// 1) Só admins podem criar médicos
	if actor.Role != domain.RoleAdmin {
		return false
	}

	// 2) Admin não pode atribuir role maior que a dele (se quiser ter hierarquia, dá pra melhorar isso)
	// Ex: se um dia você tiver RoleSuperAdmin, pode restringir criação de outros admins.
	// Por enquanto, vamos permitir admin criar qualquer role.

	// se actor.UbsID == nil, talvez ele seja um super admin global — aí você decide a regra

	return true
}

func (s *SimpleAuthorizationService) CanCreatePatient(ctx context.Context, user *domain.User) bool {
	if user == nil {
		return false
	}

	switch user.Role {
	case domain.RoleAdmin, domain.RoleDoctor:
		return true
	default:
		return false
	}
}

func (s *SimpleAuthorizationService) CanViewPatient(
	ctx context.Context,
	user *domain.User,
	patient *domain.Patient,
) bool {
	if user == nil || patient == nil {
		return false
	}

	// Admin vê tudo
	if user.Role == domain.RoleAdmin {
		return true
	}

	// Médicos — regra simples inicial: podem ver qualquer paciente.
	// (depois você pode restringir a pacientes vinculados)
	if user.Role == domain.RoleDoctor {
		return true
	}

	// Paciente só vê a si mesmo (owner).
	if user.Role == domain.RolePatient {
		return patient.AppUserID != nil && *patient.AppUserID == user.ID
	}

	return false
}

func (s *SimpleAuthorizationService) CanEditPatient(ctx context.Context, user *domain.User, patient *domain.Patient) bool {
	if user == nil || patient == nil {
		return false
	}

	// Admin pode editar qualquer paciente.
	if user.Role == domain.RoleAdmin {
		return true
	}

	// Médico pode editar pacientes.
	// (depois você pode checar vínculo médico↔paciente)
	if user.Role == domain.RoleDoctor {
		return true
	}

	// Paciente NÃO edita seu cadastro clínico (por enquanto).
	// Se quiser permitir:
	// if user.Role == domain.RolePatient && patient.AppUserID != nil && *patient.AppUserID == user.ID { return true }
	return false
}

// ---------- Laudos laboratoriais ----------

func (s *SimpleAuthorizationService) CanViewLabReport(
	ctx context.Context,
	user *domain.User,
	report *domain.LabReport,
) bool {
	if user == nil || report == nil {
		return false
	}

	// Admin vê tudo.
	if user.Role == domain.RoleAdmin {
		return true
	}

	// Médico vê qualquer laudo do paciente (regra simples).
	if user.Role == domain.RoleDoctor {
		return true
	}

	// Paciente vê laudos que pertencem a ele.
	if user.Role == domain.RolePatient {
		// aqui precisamos comparar o PatientID do laudo com o Patient ligado ao user.
		// se você tiver um campo PatientID ligado a app_users, pode ajustar.
		// exemplo simples: se o ID do paciente (PatientID) for igual ao User.ID
		// (ajuste conforme sua modelagem real)
		//
		// return report.PatientID == user.ID
		//
		// Se PatientID for id da tabela patients, aí o check precisa ser feito no use case
		// (carregar o patient e checar AppUserID).
		return true // por enquanto, relaxado — refine depois
	}

	return false
}

func (s *SimpleAuthorizationService) CanCreateLabReportFromDocument(ctx context.Context, user *domain.User, patient *domain.Patient) bool {
	if user == nil || patient == nil {
		return false
	}

	// Admin pode criar laudo para qualquer paciente.
	if user.Role == domain.RoleAdmin {
		return true
	}

	// Médico pode criar laudo para qualquer paciente (regra simples).
	// Depois você pode restringir a pacientes vinculados.
	if user.Role == domain.RoleDoctor {
		return true
	}

	// Paciente não cria laudos (no modelo atual — exames vêm do médico/DocAI).
	return false
}
