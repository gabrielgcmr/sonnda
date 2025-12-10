// internal/core/ports/services/authorization_service.go
package services

import (
	"context"
	"sonnda-api/internal/core/domain"
)

// AuthorizationService define as regras de autorização de alto nível.
// A ideia é manter aqui operações semânticas (ver/editar paciente, laudo, etc.),
// não coisas genéricas tipo "subject/object/action".
type AuthorizationService interface {
	//User
	CanCreateDoctor(ctx context.Context, actor *domain.User, newUser *domain.User) bool

	// Pacientes
	CanViewPatient(ctx context.Context, user *domain.User, patient *domain.Patient) bool
	CanEditPatient(ctx context.Context, user *domain.User, patient *domain.Patient) bool
	CanCreatePatient(ctx context.Context, user *domain.User) bool

	// Laudos laboratoriais
	CanViewLabReport(ctx context.Context, user *domain.User, report *domain.LabReport) bool
	CanCreateLabReportFromDocument(ctx context.Context, user *domain.User, patient *domain.Patient) bool
}
