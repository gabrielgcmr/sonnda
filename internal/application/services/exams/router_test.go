// internal/application/services/exams/router_test.go
package examsvc

import (
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
)

func TestHeuristicExamRouter_RoutesLaboratoryReport(t *testing.T) {
	router := NewHeuristicExamRouter()

	result := router.Route(ExamRouteInput{
		ExtractedText: "Hemograma completo\nLeucócitos 7200\nPlaquetas 250000\nValor de referência",
	})

	if result.ExamType != exams.ExamTypeLaboratory {
		t.Fatalf("expected laboratory, got %s", result.ExamType)
	}
	if result.Status != exams.DocumentStatusProcessed {
		t.Fatalf("expected processed, got %s", result.Status)
	}
	if result.Confidence < 0.70 {
		t.Fatalf("expected confidence >= 0.70, got %.2f", result.Confidence)
	}
}

func TestHeuristicExamRouter_RoutesImagingReport(t *testing.T) {
	router := NewHeuristicExamRouter()

	result := router.Route(ExamRouteInput{
		ExtractedText: "Ressonância magnética do joelho\nTécnica do exame\nAchados\nImpressão diagnóstica",
	})

	if result.ExamType != exams.ExamTypeImaging {
		t.Fatalf("expected imaging, got %s", result.ExamType)
	}
	if result.Status != exams.DocumentStatusProcessed {
		t.Fatalf("expected processed, got %s", result.Status)
	}
	if result.Confidence < 0.70 {
		t.Fatalf("expected confidence >= 0.70, got %.2f", result.Confidence)
	}
}

func TestHeuristicExamRouter_RoutesUltrasoundAsImagingEvenWithMeasurements(t *testing.T) {
	router := NewHeuristicExamRouter()

	result := router.Route(ExamRouteInput{
		ExtractedText: "ULTRASSONOGRAFIA DO APARELHO URINARIO\nAnalise: calculo renal medindo 0,5 cm. Cisto simples medindo 2,23 x 2,65 x 2,12 cm.",
	})

	if result.ExamType != exams.ExamTypeImaging {
		t.Fatalf("expected imaging, got %s", result.ExamType)
	}
}

func TestHeuristicExamRouter_RoutesMedicalCertificateAsUnknown(t *testing.T) {
	router := NewHeuristicExamRouter()

	result := router.Route(ExamRouteInput{
		ExtractedText: "ATESTADO MEDICO\nPaciente devera ficar afastado de suas atividades por 04 dias. CID: J18.9",
	})

	if result.ExamType != exams.ExamTypeUnknown {
		t.Fatalf("expected unknown, got %s", result.ExamType)
	}
	if result.Status != exams.DocumentStatusNeedsReview {
		t.Fatalf("expected needs_review, got %s", result.Status)
	}
}

func TestHeuristicExamRouter_RoutesExamRequestAsUnknownEvenWithLabNames(t *testing.T) {
	router := NewHeuristicExamRouter()

	result := router.Route(ExamRouteInput{
		ExtractedText: "Pedido de exames\nSolicito: hemograma completo, glicemia de jejum e creatinina.",
	})

	if result.ExamType != exams.ExamTypeUnknown {
		t.Fatalf("expected unknown, got %s", result.ExamType)
	}
}

func TestHeuristicExamRouter_RoutesUnknownWhenAmbiguous(t *testing.T) {
	router := NewHeuristicExamRouter()

	result := router.Route(ExamRouteInput{
		ExtractedText: "Exame recebido para avaliação posterior",
	})

	if result.ExamType != exams.ExamTypeUnknown {
		t.Fatalf("expected unknown, got %s", result.ExamType)
	}
	if result.Status != exams.DocumentStatusNeedsReview {
		t.Fatalf("expected needs_review, got %s", result.Status)
	}
}

func TestHeuristicExamRouter_UsesFilenameAsSignal(t *testing.T) {
	router := NewHeuristicExamRouter()

	result := router.Route(ExamRouteInput{
		OriginalFilename: "tomografia-abdome.pdf",
	})

	if result.ExamType != exams.ExamTypeImaging {
		t.Fatalf("expected imaging, got %s", result.ExamType)
	}
	if result.Status != exams.DocumentStatusNeedsReview {
		t.Fatalf("expected needs_review for filename-only match, got %s", result.Status)
	}
}

func TestHeuristicExamRouter_UsesLaboratoryFilenameAsSignal(t *testing.T) {
	router := NewHeuristicExamRouter()

	result := router.Route(ExamRouteInput{
		OriginalFilename: "hemograma-completo.pdf",
	})

	if result.ExamType != exams.ExamTypeLaboratory {
		t.Fatalf("expected laboratory, got %s", result.ExamType)
	}
	if result.Status != exams.DocumentStatusNeedsReview {
		t.Fatalf("expected needs_review for filename-only match, got %s", result.Status)
	}
}
