// internal/api/handlers/medicalrecord/exams.go
package medicalrecord

import (
	base "github.com/gabrielgcmr/sonnda/internal/api/handlers"
	authorization "github.com/gabrielgcmr/sonnda/internal/application/services/authorization"
	examsvc "github.com/gabrielgcmr/sonnda/internal/application/services/exams"
	textsvc "github.com/gabrielgcmr/sonnda/internal/application/services/textextraction"
	labsuc "github.com/gabrielgcmr/sonnda/internal/application/usecase/labs"
	domainstorage "github.com/gabrielgcmr/sonnda/internal/domain/storage"
)

type ExamsHandler = base.ExamsHandler

func NewExamsHandler(
	examService examsvc.Service,
	createLabUC labsuc.CreateLabReportFromDocumentUseCase,
	storage domainstorage.FileStorageService,
	textExtractor textsvc.Extractor,
	authz authorization.Authorizer,
) *ExamsHandler {
	return base.NewExams(examService, createLabUC, storage, textExtractor, authz)
}
