// internal/features/documentprocessing/postgres/repository_integration_test.go
//go:build integration

package postgres

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	processingdomain "github.com/gabrielgcmr/sonnda/internal/features/documentprocessing/domain"
	labs "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/domain"
	labpostgres "github.com/gabrielgcmr/sonnda/internal/features/patient/exam/laboratory/postgres"
	pginfra "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func processingTestDatabase(t *testing.T) (*pginfra.Client, *Repository, uuid.UUID, uuid.UUID) {
	t.Helper()
	rawURL := os.Getenv("LABS_TEST_DATABASE_URL")
	if rawURL == "" {
		t.Skip("set LABS_TEST_DATABASE_URL to an isolated local Postgres")
	}
	u, err := url.Parse(rawURL)
	if err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") {
		t.Fatal("integration tests require a local Postgres URL")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, rawURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close(ctx) })
	schema := "processing_test_" + uuid.NewString()[:8]
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	params := u.Query()
	params.Set("search_path", schema)
	u.RawQuery = params.Encode()
	client, err := pginfra.NewClient(pginfra.Config{DatabaseURL: u.String(), MaxConns: 4})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	if _, err := client.Pool().Exec(ctx, "CREATE TABLE users (id uuid PRIMARY KEY); CREATE TABLE patients (id uuid PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	schemaDir := filepath.Join("..", "..", "..", "infrastructure", "persistence", "postgres", "sqlc", "sql", "schema")
	for _, name := range []string{"exam.sql", "lab.sql"} {
		sql, err := os.ReadFile(filepath.Join(schemaDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Pool().Exec(ctx, string(sql)); err != nil {
			t.Fatal(err)
		}
	}
	patientID, userID := uuid.New(), uuid.New()
	if _, err := client.Pool().Exec(ctx, "INSERT INTO users VALUES ($1); INSERT INTO patients VALUES ($2)", userID, patientID); err != nil {
		t.Fatal(err)
	}
	clinicalRepository := labpostgres.NewLabsRepository(client)
	return client, NewRepository(client, clinicalRepository), patientID, userID
}

func processingTestDocument(t *testing.T, client *pginfra.Client, patientID, userID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := client.Pool().Exec(context.Background(), `INSERT INTO exam_documents
		(id, patient_id, uploaded_by_user_id, storage_uri, original_filename, mime_type, status)
		VALUES ($1, $2, $3, 'test://exam.pdf', 'exam.pdf', 'application/pdf', 'uploaded')`, id, patientID, userID)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func processingTestReport(patientID, userID uuid.UUID) *labs.LabReport {
	reportID, resultID := uuid.New(), uuid.New()
	value, unit := "99", "mg/dL"
	return &labs.LabReport{
		ID: reportID, PatientID: patientID, UploadedBy: userID,
		TestResults: []labs.LabResult{{
			ID: resultID, LabReportID: reportID, TestName: "Glicose",
			Items: []labs.LabResultItem{{
				ID: uuid.New(), LabResultID: resultID, ParameterName: "Glicose", ResultValue: &value, ResultUnit: &unit,
			}},
		}},
	}
}

func TestCreateFromDocumentPersistsMetadataAndClinicalReport(t *testing.T) {
	client, repo, patientID, userID := processingTestDatabase(t)
	documentID := processingTestDocument(t, client, patientID, userID)
	report := processingTestReport(patientID, userID)
	rawText, fingerprint := "Hemoglobina 15,1", "fingerprint-1"
	err := repo.CreateFromDocument(context.Background(), report, processingdomain.LaboratoryReportMetadata{
		RawText: &rawText, Fingerprint: &fingerprint, ExamDocumentID: &documentID,
	})
	if err != nil {
		t.Fatal(err)
	}

	var storedRawText, storedFingerprint string
	var storedDocumentID uuid.UUID
	if err := client.Pool().QueryRow(context.Background(), `SELECT raw_text, fingerprint, exam_document_id
		FROM lab_reports WHERE id = $1`, report.ID).Scan(&storedRawText, &storedFingerprint, &storedDocumentID); err != nil {
		t.Fatal(err)
	}
	if storedRawText != rawText || storedFingerprint != fingerprint || storedDocumentID != documentID {
		t.Fatalf("processing metadata = (%q, %q, %s)", storedRawText, storedFingerprint, storedDocumentID)
	}

	stored, err := repo.FindByFingerprint(context.Background(), patientID, fingerprint)
	if err != nil || stored == nil || stored.ExamDocumentID == nil || *stored.ExamDocumentID != documentID || len(stored.TestResults) != 1 {
		t.Fatalf("processed report not returned with clinical data and document link: report=%v err=%v", stored, err)
	}
}

func TestFingerprintConflictDoesNotLeaveClinicalReport(t *testing.T) {
	client, repo, patientID, userID := processingTestDatabase(t)
	fingerprint := "duplicate-fingerprint"
	metadata := processingdomain.LaboratoryReportMetadata{Fingerprint: &fingerprint}
	original := processingTestReport(patientID, userID)
	if err := repo.CreateFromDocument(context.Background(), original, metadata); err != nil {
		t.Fatal(err)
	}

	duplicate := processingTestReport(patientID, userID)
	err := repo.CreateFromDocument(context.Background(), duplicate, metadata)
	if !errors.Is(err, labs.ErrLabReportAlreadyExists) {
		t.Fatalf("expected fingerprint conflict, got %v", err)
	}
	stored, err := repo.FindByID(context.Background(), duplicate.ID)
	if err != nil || stored != nil {
		t.Fatalf("duplicate clinical report was not cleaned up: report=%v err=%v", stored, err)
	}
	var reportCount int
	if err := client.Pool().QueryRow(context.Background(), "SELECT count(*) FROM lab_reports").Scan(&reportCount); err != nil {
		t.Fatal(err)
	}
	if reportCount != 1 {
		t.Fatalf("expected only original report to remain, got %d", reportCount)
	}
}

func TestDocumentLinkRejectsOtherPatients(t *testing.T) {
	client, repo, patientID, userID := processingTestDatabase(t)
	report := processingTestReport(patientID, userID)
	if err := repo.Create(context.Background(), report); err != nil {
		t.Fatal(err)
	}
	otherPatient := uuid.New()
	if _, err := client.Pool().Exec(context.Background(), "INSERT INTO patients VALUES ($1)", otherPatient); err != nil {
		t.Fatal(err)
	}
	documentID := processingTestDocument(t, client, otherPatient, userID)
	if err := repo.AttachDocument(context.Background(), report.ID, patientID, documentID); !errors.Is(err, labs.ErrDocumentLinkConflict) {
		t.Fatalf("cross-patient attachment accepted: %v", err)
	}
}
