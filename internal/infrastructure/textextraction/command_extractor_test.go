// internal/infrastructure/textextraction/command_extractor_test.go
package textextraction

import "testing"

func TestScoreOCRTextPrefersClinicalLabText(t *testing.T) {
	noisy := "meee eee rr iii eT an BOBIMDFTIOS 35"
	lab := "HEMOGRAMA COMPLETO Paciente OTAVIANO HEMOGLOBINA 14,6 g/dl PLAQUETAS 199000 /mm3"

	if scoreOCRText(lab) <= scoreOCRText(noisy) {
		t.Fatal("expected lab text to score higher than noisy OCR text")
	}
}
