// internal/core/ports/services/authorization_service.go
package services

import (
	"context"
	"sonnda-api/internal/core/domain/identity"
	"sonnda-api/internal/core/domain/medicalRecord/exam/lab"
	"sonnda-api/internal/core/domain/patient"
)

// AuthorizationService define as regras de autorização de alto nível.
// A ideia é manter aqui operações semânticas (ver/editar paciente, laudo, etc.),
// não coisas genéricas tipo "subject/object/action".
type AuthorizationService interface {
	//User
	CanCreateDoctor(ctx context.Context, actor *identity.User, newUser *identity.User) bool

	// Pacientes
	CanViewPatient(ctx context.Context, user *identity.User, patient *patient.Patient) bool
	CanEditPatient(ctx context.Context, user *identity.User, patient *patient.Patient) bool
	CanCreatePatient(ctx context.Context, user *identity.User) bool

	// Laudos laboratoriais
	CanViewLabReport(ctx context.Context, user *identity.User, report *lab.LabReport) bool
	CanCreateLabReportFromDocument(ctx context.Context, user *identity.User, patient *patient.Patient) bool
}
