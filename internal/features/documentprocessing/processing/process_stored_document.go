// internal/features/documentprocessing/processing/process_stored_document.go
package processing

import (
	"context"
	"errors"
	"strings"
	"time"

	domaintext "github.com/gabrielgcmr/sonnda/internal/domain/textextraction"
	documents "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing"
	documentdomain "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type ProcessStoredDocumentUseCase interface {
	Execute(ctx context.Context, input ProcessStoredDocumentInput) (*documents.ExamDocumentOutput, error)
}

type ProcessStoredDocumentInput struct {
	PatientID        uuid.UUID
	UploadedByUserID uuid.UUID
	StorageURI       string
	OriginalFilename string
	MimeType         string
	LocalPath        string
	CollectionDate   *time.Time
}

type processStoredDocumentUseCase struct {
	documents  documents.Service
	extractor  domaintext.Extractor
	processors map[DocumentKind]Processor
}

var _ ProcessStoredDocumentUseCase = (*processStoredDocumentUseCase)(nil)

func NewProcessStoredDocument(
	documentService documents.Service,
	extractor domaintext.Extractor,
	processors ...Processor,
) ProcessStoredDocumentUseCase {
	registered := make(map[DocumentKind]Processor, len(processors))
	for _, processor := range processors {
		if processor != nil {
			registered[processor.Kind()] = processor
		}
	}
	return &processStoredDocumentUseCase{
		documents:  documentService,
		extractor:  extractor,
		processors: registered,
	}
}

func (u *processStoredDocumentUseCase) Execute(ctx context.Context, input ProcessStoredDocumentInput) (*documents.ExamDocumentOutput, error) {
	document, err := u.documents.Create(ctx, documents.CreateExamDocumentInput{
		PatientID:        input.PatientID,
		UploadedByUserID: input.UploadedByUserID,
		StorageURI:       input.StorageURI,
		OriginalFilename: input.OriginalFilename,
		MimeType:         input.MimeType,
	})
	if err != nil {
		return nil, err
	}

	routed, processingErr := u.extractAndRoute(ctx, document.ID, input)
	if routed == nil {
		routed = u.routeFromMetadata(ctx, document.ID, processingErr)
	}
	if routed == nil {
		return document, nil
	}

	document = routed.document
	if routed.extracted == nil {
		return document, nil
	}
	if document.Status == documentdomain.DocumentStatusNeedsReview {
		u.createDocumentTextFromExtraction(ctx, document, routed.extracted, input.CollectionDate)
		return document, nil
	}
	if document.ExamType == nil || *document.ExamType == DocumentKindUnknown {
		reviewed, err := u.markNeedsReview(ctx, document, routed.extracted, "Nao foi possivel identificar o tipo de documento com seguranca.")
		if err != nil {
			return nil, err
		}
		u.createDocumentTextFromExtraction(ctx, reviewed, routed.extracted, input.CollectionDate)
		return reviewed, nil
	}

	processor := u.processors[DocumentKind(*document.ExamType)]
	if processor == nil {
		reviewed, err := u.markNeedsReview(ctx, document, routed.extracted, "Nao existe processador configurado para este tipo de documento.")
		if err != nil {
			return nil, err
		}
		u.createDocumentTextFromExtraction(ctx, reviewed, routed.extracted, input.CollectionDate)
		return reviewed, nil
	}

	if err := processor.Process(ctx, ProcessingInput{
		Document:       document,
		ExtractedText:  routed.extracted,
		CollectionDate: input.CollectionDate,
	}); err != nil {
		if errors.Is(err, ErrProcessorNeedsReview) {
			reviewed, reviewErr := u.markNeedsReview(ctx, document, routed.extracted, "O documento nao produziu resultados clinicos estruturados e precisa de revisao.")
			if reviewErr != nil {
				return nil, reviewErr
			}
			u.createDocumentTextFromExtraction(ctx, reviewed, routed.extracted, input.CollectionDate)
			return reviewed, nil
		}
		return nil, u.markProcessorFailure(ctx, document.ID, err)
	}

	return document, nil
}

type routedDocument struct {
	document  *documents.ExamDocumentOutput
	extracted *domaintext.ExtractOutput
}

func (u *processStoredDocumentUseCase) extractAndRoute(
	ctx context.Context,
	documentID uuid.UUID,
	input ProcessStoredDocumentInput,
) (*routedDocument, *apperr.AppError) {
	if u.extractor == nil {
		return nil, apperr.Internal("A leitura automatica esta indisponivel. O arquivo foi salvo e precisa de revisao.", nil)
	}

	extracted, err := u.extractor.Extract(ctx, domaintext.ExtractInput{
		LocalPath:        input.LocalPath,
		DocumentURI:      input.StorageURI,
		MimeType:         input.MimeType,
		OriginalFilename: input.OriginalFilename,
	})
	if err != nil {
		message := "Nao foi possivel ler o texto do documento. Tente enviar uma foto mais nitida ou o PDF original."
		var appErr *apperr.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		if errors.Is(err, context.DeadlineExceeded) {
			message = "A leitura do documento excedeu o tempo limite. O arquivo foi salvo. Tente enviar novamente ou usar o PDF original."
		}
		return nil, apperr.Internal(message, err)
	}
	if extracted == nil || strings.TrimSpace(extracted.Text) == "" {
		return nil, apperr.Internal("Nao foi encontrado texto legivel no documento. Tente enviar uma foto mais nitida ou o PDF original.", nil)
	}

	document, err := u.documents.RouteDocument(ctx, documents.RouteExamDocumentInput{
		ID:               documentID,
		ExtractedText:    extracted.Text,
		ExtractionMethod: extracted.Method,
	})
	if err != nil {
		return nil, apperr.Internal("Nao foi possivel concluir a classificacao do exame. O arquivo foi salvo e precisa de revisao.", err)
	}
	if document == nil {
		return nil, apperr.Internal("Nao foi possivel concluir a classificacao do exame. O arquivo foi salvo e precisa de revisao.", nil)
	}

	return &routedDocument{document: document, extracted: extracted}, nil
}

func (u *processStoredDocumentUseCase) routeFromMetadata(
	ctx context.Context,
	documentID uuid.UUID,
	processingErr *apperr.AppError,
) *routedDocument {
	document, err := u.documents.RouteDocument(ctx, documents.RouteExamDocumentInput{
		ID:               documentID,
		ExtractionMethod: "metadata",
		ProcessingError:  processingErr,
	})
	if err != nil || document == nil {
		return nil
	}
	return &routedDocument{document: document}
}

func (u *processStoredDocumentUseCase) markNeedsReview(
	ctx context.Context,
	document *documents.ExamDocumentOutput,
	extracted *domaintext.ExtractOutput,
	message string,
) (*documents.ExamDocumentOutput, error) {
	input := documents.RouteExamDocumentInput{
		ID:               document.ID,
		ExtractionMethod: optionalExtractionMethod(extracted),
		ProcessingError:  apperr.Internal(message, nil),
	}
	if extracted != nil {
		input.ExtractedText = extracted.Text
	}
	reviewed, err := u.documents.RouteDocument(ctx, input)
	if err != nil {
		return nil, apperr.Internal("falha ao registrar revisao do documento", err)
	}
	if reviewed == nil {
		return nil, apperr.NotFound("documento nao encontrado")
	}
	return reviewed, nil
}

func (u *processStoredDocumentUseCase) markProcessorFailure(ctx context.Context, documentID uuid.UUID, processingErr error) error {
	message := "falha no processamento do documento"
	var appErr *apperr.AppError
	if errors.As(processingErr, &appErr) {
		message = appErr.Message
	}

	if _, err := u.documents.MarkFailed(ctx, documents.MarkExamDocumentFailedInput{ID: documentID, ErrorMessage: message}); err != nil {
		return apperr.Internal("falha ao registrar erro do exame", errors.Join(processingErr, err))
	}
	return processingErr
}

func (u *processStoredDocumentUseCase) createDocumentTextFromExtraction(
	ctx context.Context,
	document *documents.ExamDocumentOutput,
	extracted *domaintext.ExtractOutput,
	collectionDate *time.Time,
) {
	if document == nil || extracted == nil {
		return
	}

	category := documentdomain.ExamTypeUnknown
	if document.ExamType != nil {
		category = *document.ExamType
	}
	_, _ = u.documents.CreateDocumentTextFromText(ctx, documents.CreateExamDocumentTextFromTextInput{
		ExamDocumentID:   document.ID,
		PatientID:        document.PatientID,
		UploadedByUserID: document.UploadedByUserID,
		Category:         category,
		Text:             documentTextForDisplay(extracted),
		PerformedAt:      collectionDate,
		ExtractionMethod: extracted.Method,
		Confidence:       document.Confidence,
	})
}

func documentTextForDisplay(extracted *domaintext.ExtractOutput) string {
	if extracted == nil || strings.TrimSpace(extracted.NormalizedText) == "" {
		if extracted == nil {
			return ""
		}
		return extracted.Text
	}
	return extracted.NormalizedText
}

func optionalExtractionMethod(extracted *domaintext.ExtractOutput) string {
	if extracted == nil {
		return "metadata"
	}
	return extracted.Method
}
