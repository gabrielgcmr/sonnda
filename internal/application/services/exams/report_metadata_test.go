package examsvc

import (
	"strings"
	"testing"
)

func TestInferReportMetadata_UrinaryUltrasound(t *testing.T) {
	text := `CENTROMEB

ULTRASSONOGRAFIA DO APARELHO URINARIO

METODOLOGIA:
Exame realizado com transdutor multifrequencial.

ANALISE:
Rins topicos com forma, dimensoes e contornos habituais.

OPINIAO:
- Achados sugestivos de nefrolitiase nao obstrutiva a direita.
- Cisto renal a direita.`

	metadata := inferReportMetadata(text)

	if metadata.Title == nil || *metadata.Title != "ULTRASSONOGRAFIA DO APARELHO URINARIO" {
		t.Fatalf("unexpected title: %v", metadata.Title)
	}
	if metadata.Modality == nil || *metadata.Modality != "ultrasound" {
		t.Fatalf("unexpected modality: %v", metadata.Modality)
	}
	if metadata.BodySite == nil || *metadata.BodySite != "aparelho_urinario" {
		t.Fatalf("unexpected body site: %v", metadata.BodySite)
	}
	if metadata.Conclusion == nil || !strings.Contains(*metadata.Conclusion, "nefrolitiase") {
		t.Fatalf("unexpected conclusion: %v", metadata.Conclusion)
	}
}
