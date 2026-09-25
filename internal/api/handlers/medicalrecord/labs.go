// internal/api/handlers/medicalrecord/labs.go
package medicalrecord

import (
	base "github.com/gabrielgcmr/sonnda/internal/api/handlers"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	labsvc "github.com/gabrielgcmr/sonnda/internal/application/services/labs"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
)

type LabsHandler = base.LabsHandler

func NewLabsHandler(
	svc labsvc.Service,
	createUC labsuc.CreateLabReportFromDocumentUseCase,
	storageClient domainstorage.FileStorageService,
	authz authorization.Authorizer,
) *LabsHandler {
	return base.NewLabs(svc, createUC, storageClient, authz)
}
