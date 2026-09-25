// internal/api/handlers/medicalrecord/exams.go
package medicalrecord

import (
	base "github.com/gabrielgcmr/sonnda/internal/api/handlers"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
)

type ExamsHandler = base.ExamsHandler

func NewExamsHandler(
	examService examsvc.Service,
	createLabUC labsuc.CreateLabReportFromDocumentUseCase,
	storage domainstorage.FileStorageService,
	textExtractor domaintext.Extractor,
	authz authorization.Authorizer,
) *ExamsHandler {
	return base.NewExams(examService, createLabUC, storage, textExtractor, authz)
}
