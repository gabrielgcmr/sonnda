package integrations

import (
	"context"

	"sonnda-api/internal/domain/entities/medicalrecord/lab"
	"sonnda-api/internal/domain/entities/patient"
	"sonnda-api/internal/domain/entities/user"
)

// AuthorizationService define as regras de autorização de alto nível.
// A ideia é manter aqui operações semânticas (ver/editar paciente, laudo, etc.),
// não coisas genéricas tipo "subject/object/action".
type AuthorizationService interface {
	//User
	CanCreateDoctor(ctx context.Context, actor *user.User, newUser *user.User) bool

	// Pacientes
	CanViewPatient(ctx context.Context, user *user.User, patient *patient.Patient) bool
	CanEditPatient(ctx context.Context, user *user.User, patient *patient.Patient) bool
	CanCreatePatient(ctx context.Context, user *user.User) bool

	// Laudos laboratoriais
	CanViewLabReport(ctx context.Context, user *user.User, report *lab.LabReport) bool
	CanCreateLabReportFromDocument(ctx context.Context, user *user.User, patient *patient.Patient) bool
}
