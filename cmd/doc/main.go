package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	documentai "cloud.google.com/go/documentai/apiv1"
	documentaipb "cloud.google.com/go/documentai/apiv1/documentaipb"
)

type LabReport struct {
	ID        string    `db:"id"          json:"id"`
	PatientID string    `db:"patient_id"  json:"patient_id"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`

	// Metadados extraídos do cabeçalho do laudo
	PatientName       *string    `db:"patient_name"       json:"patient_name,omitempty"`
	PatientDOB        *time.Time `db:"patient_dob"        json:"patient_dob,omitempty"`
	LabName           *string    `db:"lab_name"           json:"lab_name,omitempty"`
	LabPhone          *string    `db:"lab_phone"          json:"lab_phone,omitempty"`
	InsuranceProvider *string    `db:"insurance_provider" json:"insurance_provider,omitempty"`
	RequestingDoctor  *string    `db:"requesting_doctor"  json:"requesting_doctor,omitempty"`
	TechnicalManager  *string    `db:"technical_manager"  json:"technical_manager,omitempty"`
	ReportDate        *time.Time `db:"report_date"        json:"report_date,omitempty"`

	RawText *string `db:"raw_text" json:"raw_text,omitempty"`

	// Carregado via JOIN quando você quiser devolver tudo de uma vez
	TestResults []LabResult `json:"test_results,omitempty"`
}

// LabResult representa um exame/painel dentro do laudo
// (ex.: Hemograma, Creatinina, HbA1c). Uma linha em lab_results.
type LabResult struct {
	ID          string `db:"id"            json:"id"`
	LabReportID string `db:"lab_report_id" json:"lab_report_id"`

	TestName string  `db:"test_name" json:"test_name"`
	Material *string `db:"material"  json:"material,omitempty"`
	Method   *string `db:"method"    json:"method,omitempty"`

	CollectedAt *time.Time `db:"collected_at" json:"collected_at,omitempty"`
	ReleaseAt   *time.Time `db:"release_at"   json:"release_at,omitempty"`

	// Carregado via JOIN quando necessário
	Items []LabResultItem `json:"items,omitempty"`
}

// LabResultItem representa uma linha/paramêtro dentro de um teste
// (ex.: Hemoglobina, Creatinina, LDL). Uma linha em lab_result_items.
type LabResultItem struct {
	ID          string `db:"id"                 json:"id"`
	LabResultID string `db:"lab_result_id" json:"lab_result_id"`

	ParameterName string  `db:"parameter_name" json:"parameter_name"`
	ResultValue   *string `db:"result_value"    json:"result_value,omitempty"`
	ResultUnit    *string `db:"result_unit"    json:"result_unit,omitempty"`
	ReferenceText *string `db:"reference_text" json:"reference_text,omitempty"`
}

// --- FUNÇÕES AUXILIARES PARA CONVERSÃO ---

// stringPtr converte string para *string (útil para campos opcionais)
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// parseDate tenta converter string (dd/mm/yyyy) para *time.Time
// Nota: O formato OCR pode variar, aqui assumimos o padrão BR do seu PDF.
func parseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	// Tenta formato DD/MM/YYYY
	layout := "02/01/2006"
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		// Se falhar, tenta formato com hora DD/MM/YYYY HH:MM
		layoutTime := "02/01/2006 15:04"
		t, err = time.Parse(layoutTime, dateStr)
		if err != nil {
			// Se falhar tudo, retorna nil (para não quebrar a aplicação)
			// Idealmente, logar o erro ou tentar limpar a string
			return nil
		}
	}
	return &t
}

// --- LÓGICA PRINCIPAL ---

func ProcessLabReport(projectID, location, processorID, filePath string) (*LabReport, error) {
	ctx := context.Background()

	c, err := documentai.NewDocumentProcessorClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente: %v", err)
	}
	defer c.Close()

	// CORREÇÃO 1: Usando os.ReadFile no lugar de ioutil.ReadFile
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	name := fmt.Sprintf("projects/%s/locations/%s/processors/%s", projectID, location, processorID)

	req := &documentaipb.ProcessRequest{
		Name: name,
		Source: &documentaipb.ProcessRequest_RawDocument{
			RawDocument: &documentaipb.RawDocument{
				Content:  fileData,
				MimeType: "application/pdf",
			},
		},
	}

	resp, err := c.ProcessDocument(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("falha no processamento do DocAI: %v", err)
	}

	// Inicializa o Report Raiz
	report := &LabReport{
		RawText: stringPtr(resp.Document.Text),
	}

	// Varre todas as entidades encontradas
	for _, entity := range resp.Document.Entities {

		// Normaliza o texto removendo quebras de linha extras se necessário
		textValue := entity.MentionText

		switch entity.Type {
		// --- Mapeamento do Cabeçalho (Nível Report) ---
		case "patient_name":
			report.PatientName = stringPtr(textValue)
		case "patient_dob":
			report.PatientDOB = parseDate(textValue)
		case "lab_name":
			report.LabName = stringPtr(textValue)
		case "lab_phone":
			report.LabPhone = stringPtr(textValue)
		case "insurance_provider":
			report.InsuranceProvider = stringPtr(textValue)
		case "requesting_doctor":
			report.RequestingDoctor = stringPtr(textValue)
		case "technical_manager":
			report.TechnicalManager = stringPtr(textValue)
		case "report_date": // Data de emissão do laudo geral
			report.ReportDate = parseDate(textValue)

		// --- Mapeamento dos Exames (Nível TestResult) ---
		case "test_result":
			testResult := LabResult{}

			// Itera sobre as propriedades dentro do test_result
			for _, prop := range entity.Properties {
				propText := prop.MentionText

				switch prop.Type {
				case "test_name":
					testResult.TestName = propText
				case "collection_datetime":
					testResult.CollectedAt = parseDate(propText)
				case "release_datetime":
					testResult.ReleaseAt = parseDate(propText)
				case "material":
					testResult.Material = stringPtr(propText)
				case "method":
					testResult.Method = stringPtr(propText)

				// --- Mapeamento dos Itens (Nível TestItem) ---
				case "test_item":
					item := LabResultItem{}
					for _, subProp := range prop.Properties {
						subText := subProp.MentionText
						switch subProp.Type {
						case "parameter_name":
							item.ParameterName = subText
						case "result": // mapeado para result_value no DocAI
							item.ResultValue = stringPtr(subText)
						case "unit":
							item.ResultUnit = stringPtr(subText)
						case "reference_text":
							item.ReferenceText = stringPtr(subText)
						}
					}
					// Adiciona o item ao resultado do teste
					testResult.Items = append(testResult.Items, item)
				}
			}
			// Adiciona o resultado do teste ao relatório
			report.TestResults = append(report.TestResults, testResult)
		}
	}

	return report, nil
}

func main() {
	// Exemplo de execução (Altere com suas credenciais REAIS ou variáveis de ambiente)
	projectID := "766877120359"
	location := "us" // ou "us-central1"
	processorID := "d07a2bfe9bb83a01"
	filePath := "C:\\Users\\gabri\\Dev\\sonnda-api\\labtest.pdf"

	// Necessário configurar autenticação antes
	// os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "path/to/key.json")

	report, err := ProcessLabReport(projectID, location, processorID, filePath)
	if err != nil {
		log.Fatalf("Erro: %v", err)
	}

	fmt.Printf("Relatório processado para: %v\n", *report.PatientName)
	for _, result := range report.TestResults {
		fmt.Printf(" >> Exame: %s (Itens: %d)\n", result.TestName, len(result.Items))
		for _, item := range result.Items {
			val := "N/A"
			if item.ResultValue != nil {
				val = *item.ResultValue
			}
			fmt.Printf("    - %s: %s\n", item.ParameterName, val)
		}
	}
}
