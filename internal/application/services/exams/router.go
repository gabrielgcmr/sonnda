// internal/application/services/exams/router.go
package examsvc

import (
	"math"
	"strings"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/exams"
)

type ExamRouter interface {
	Route(input ExamRouteInput) ExamRouteResult
}

type ExamRouteInput struct {
	ExtractedText    string
	MimeType         string
	OriginalFilename string
}

type ExamRouteResult struct {
	ExamType       exams.ExamType
	Status         exams.DocumentStatus
	Confidence     float64
	MatchedSignals []string
}

type HeuristicExamRouter struct{}

func NewHeuristicExamRouter() *HeuristicExamRouter {
	return &HeuristicExamRouter{}
}

func (r *HeuristicExamRouter) Route(input ExamRouteInput) ExamRouteResult {
	hasExtractedText := strings.TrimSpace(input.ExtractedText) != ""
	text := normalizeRouteText(strings.Join([]string{
		input.ExtractedText,
		input.OriginalFilename,
		input.MimeType,
	}, " "))

	labScore, labSignals := scoreSignals(text, laboratorySignals)
	imagingScore, imagingSignals := scoreSignals(text, imagingSignals)
	nonResultScore, nonResultSignals := scoreSignals(text, nonLaboratoryResultSignals)

	if nonResultScore > 0 && !hasLaboratoryResultEvidence(text) {
		confidence := 0.45
		if hasExtractedText {
			confidence = 0.65
		}
		return ExamRouteResult{
			ExamType:       exams.ExamTypeUnknown,
			Status:         exams.DocumentStatusNeedsReview,
			Confidence:     confidence,
			MatchedSignals: nonResultSignals,
		}
	}

	if labScore == 0 && imagingScore == 0 {
		return ExamRouteResult{
			ExamType:       exams.ExamTypeUnknown,
			Status:         exams.DocumentStatusNeedsReview,
			Confidence:     0.20,
			MatchedSignals: nil,
		}
	}

	if imagingScore > labScore {
		confidence := confidenceForScores(imagingScore, labScore, hasExtractedText)
		status := statusForConfidence(confidence)
		return ExamRouteResult{
			ExamType:       exams.ExamTypeImaging,
			Status:         status,
			Confidence:     confidence,
			MatchedSignals: imagingSignals,
		}
	}

	if labScore > imagingScore {
		confidence := confidenceForScores(labScore, imagingScore, hasExtractedText)
		status := statusForConfidence(confidence)
		return ExamRouteResult{
			ExamType:       exams.ExamTypeLaboratory,
			Status:         status,
			Confidence:     confidence,
			MatchedSignals: labSignals,
		}
	}

	return ExamRouteResult{
		ExamType:       exams.ExamTypeUnknown,
		Status:         exams.DocumentStatusNeedsReview,
		Confidence:     0.45,
		MatchedSignals: append(labSignals, imagingSignals...),
	}
}

type weightedSignal struct {
	term   string
	weight int
}

var laboratorySignals = []weightedSignal{
	{term: "hemograma", weight: 4},
	{term: "leucocitos", weight: 3},
	{term: "hemacias", weight: 3},
	{term: "plaquetas", weight: 3},
	{term: "glicose", weight: 3},
	{term: "creatinina", weight: 3},
	{term: "ureia", weight: 2},
	{term: "colesterol", weight: 3},
	{term: "triglicerides", weight: 3},
	{term: "hdl", weight: 2},
	{term: "ldl", weight: 2},
	{term: "tsh", weight: 2},
	{term: "t4 livre", weight: 2},
	{term: "valor de referencia", weight: 4},
	{term: "valores de referencia", weight: 4},
	{term: "material: sangue", weight: 3},
	{term: "material soro", weight: 2},
	{term: "data coleta", weight: 2},
	{term: "data da coleta", weight: 2},
	{term: "metodo:", weight: 1},
	{term: "resultado unidade referencia", weight: 4},
}

var imagingSignals = []weightedSignal{
	{term: "tomografia", weight: 4},
	{term: "ressonancia", weight: 4},
	{term: "ultrassonografia", weight: 4},
	{term: "ultrassom", weight: 3},
	{term: "radiografia", weight: 4},
	{term: "raio x", weight: 4},
	{term: "raios x", weight: 4},
	{term: "mamografia", weight: 4},
	{term: "ecografia", weight: 3},
	{term: "densitometria", weight: 3},
	{term: "achados", weight: 3},
	{term: "impressao diagnostica", weight: 5},
	{term: "conclusao", weight: 2},
	{term: "tecnica do exame", weight: 3},
	{term: "contraste", weight: 2},
	{term: "radiologista", weight: 3},
	{term: "laudo radiologico", weight: 5},
	{term: "cortes axiais", weight: 4},
	{term: "sequencias ponderadas", weight: 4},
}

var nonLaboratoryResultSignals = []weightedSignal{
	{term: "atestado medico", weight: 5},
	{term: "afastado de suas atividades", weight: 5},
	{term: "cid:", weight: 4},
	{term: "pedido de exame", weight: 4},
	{term: "pedido de exames", weight: 4},
	{term: "solicitacao de exame", weight: 4},
	{term: "solicitacao de exames", weight: 4},
	{term: "requisicao de exame", weight: 4},
	{term: "requisicao de exames", weight: 4},
	{term: "solicito:", weight: 3},
	{term: "prescricao", weight: 3},
}

var laboratoryResultEvidenceSignals = []weightedSignal{
	{term: "valor de referencia", weight: 4},
	{term: "valores de referencia", weight: 4},
	{term: "resultado unidade referencia", weight: 4},
	{term: "data coleta", weight: 2},
	{term: "data da coleta", weight: 2},
	{term: "material: sangue", weight: 2},
	{term: "material sangue", weight: 2},
	{term: "material: soro", weight: 2},
	{term: "material soro", weight: 2},
	{term: "liberado em", weight: 2},
}

func scoreSignals(text string, signals []weightedSignal) (int, []string) {
	score := 0
	matched := make([]string, 0)
	for _, signal := range signals {
		if strings.Contains(text, signal.term) {
			score += signal.weight
			matched = append(matched, signal.term)
		}
	}
	return score, matched
}

func hasLaboratoryResultEvidence(text string) bool {
	score, _ := scoreSignals(text, laboratoryResultEvidenceSignals)
	return score >= 2
}

func confidenceForScores(winnerScore, loserScore int, hasExtractedText bool) float64 {
	gap := winnerScore - loserScore
	confidence := 0.55 + math.Min(float64(winnerScore), 10)*0.035 + math.Min(float64(gap), 8)*0.025
	if !hasExtractedText && confidence > 0.65 {
		return 0.65
	}
	if confidence > 0.95 {
		return 0.95
	}
	return confidence
}

func statusForConfidence(confidence float64) exams.DocumentStatus {
	if confidence < 0.70 {
		return exams.DocumentStatusNeedsReview
	}
	return exams.DocumentStatusProcessed
}

func normalizeRouteText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c",
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c",
		"-", " ", "_", " ", "/", " ", "\\", " ",
		"\n", " ", "\r", " ", "\t", " ",
	)
	text = replacer.Replace(text)
	return strings.Join(strings.Fields(text), " ")
}
