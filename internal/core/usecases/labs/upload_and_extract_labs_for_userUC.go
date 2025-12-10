package labs

import (
	"context"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"
)

type CreateFromDocumentForUserUseCase struct {
	base        *CreateFromDocumentUseCase
	patientRepo repositories.PatientRepository
}

func NewCreateFromDocumentForUser(
	base *CreateFromDocumentUseCase,
	patientRepo repositories.PatientRepository,
) *CreateFromDocumentForUserUseCase {
	return &CreateFromDocumentForUserUseCase{
		base:        base,
		patientRepo: patientRepo,
	}
}

func (uc *CreateFromDocumentForUserUseCase) Execute(
	ctx context.Context,
	currentUser *domain.User,
	input CreateFromDocumentForCurrentUserInput,
) (*LabReportOutput, error) {
	// 1) garantir que é um patient
	if currentUser.Role != domain.RolePatient {
		return nil, domain.ErrForbidden // ou erro específico
	}

	// 2) buscar o Patient ligado a esse user
	patient, err := uc.patientRepo.FindByUserID(ctx, currentUser.ID)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		// aqui você pode decidir se:
		// - retorna erro (não há patient vinculado),
		// - ou cria um Patient “on the fly” com base em dados mínimos do user.
		return nil, domain.ErrPatientNotFound
	}

	// 3) delegar pro use case original, agora com PatientID resolvido
	baseInput := CreateFromDocumentInput{
		PatientID:   patient.ID,
		DocumentURI: input.DocumentURI,
		MimeType:    input.MimeType,
	}

	return uc.base.Execute(ctx, baseInput)
}
