// internal/infrastructure/persistence/postgres/repo/labs_integration_test.go
//go:build integration

package repo

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	postgres "github.com/gabrielgcmr/sonnda/internal/infrastructure/persistence/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Cada teste usa um schema descartavel em um Postgres local explicito.
func labTestDatabase(t *testing.T) (*postgres.Client, repository.Labs, uuid.UUID, uuid.UUID) {
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
	schema := "lab_test_" + uuid.New().String()[:8]
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
	client, err := postgres.NewClient(postgres.Config{DatabaseURL: u.String(), MaxConns: 4})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	if _, err := client.Pool().Exec(ctx, "CREATE TABLE users (id uuid PRIMARY KEY); CREATE TABLE patients (id uuid PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"exam.sql", "lab.sql"} {
		sql, err := os.ReadFile(filepath.Join("..", "sqlc", "sql", "schema", name))
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
	return client, NewLabsRepository(client), patientID, userID
}

func labTestDocument(t *testing.T, client *postgres.Client, patientID, userID uuid.UUID) uuid.UUID {
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

func labTestReport(patientID, userID uuid.UUID) *labs.LabReport {
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

func TestLabsCreateTransaction(t *testing.T) {
	for _, scenario := range []string{"success", "result_failure", "item_failure"} {
		t.Run(scenario, func(t *testing.T) {
			client, repo, patientID, userID := labTestDatabase(t)
			documentID := labTestDocument(t, client, patientID, userID)
			report := labTestReport(patientID, userID)
			report.ExamDocumentID = &documentID
			switch scenario {
			case "result_failure":
				report.TestResults = append(report.TestResults, report.TestResults[0])
			case "item_failure":
				report.TestResults[0].Items = append(report.TestResults[0].Items, report.TestResults[0].Items[0])
			}
			err := repo.Create(context.Background(), report)
			if (err == nil) != (scenario == "success") {
				t.Fatalf("unexpected create result: %v", err)
			}
			for _, table := range []string{"lab_reports", "lab_results", "lab_result_items"} {
				var count int
				if err := client.Pool().QueryRow(context.Background(), "SELECT count(*) FROM "+pgx.Identifier{table}.Sanitize()).Scan(&count); err != nil {
					t.Fatal(err)
				}
				want := 0
				if scenario == "success" {
					want = 1
				}
				if count != want {
					t.Fatalf("%s: want %d rows, got %d", table, want, count)
				}
			}
			if scenario != "success" {
				// Uma tentativa posterior deve poder salvar o mesmo laudo inteiro.
				report.TestResults = report.TestResults[:1]
				report.TestResults[0].Items = report.TestResults[0].Items[:1]
				if err := repo.Create(context.Background(), report); err != nil {
					t.Fatal(err)
				}
			}
			stored, err := repo.FindByID(context.Background(), report.ID)
			if err != nil {
				t.Fatal(err)
			}
			if stored == nil || stored.ExamDocumentID == nil || *stored.ExamDocumentID != documentID || len(stored.TestResults) != 1 || len(stored.TestResults[0].Items) != 1 {
				t.Fatal("stored report is incomplete or unlinked")
			}
		})
	}
}

func TestLabsAttachDocumentConcurrent(t *testing.T) {
	client, repo, patientID, userID := labTestDatabase(t)
	report := labTestReport(patientID, userID)
	ctx := context.Background()
	if err := repo.Create(ctx, report); err != nil {
		t.Fatal(err)
	}
	documents := []uuid.UUID{labTestDocument(t, client, patientID, userID), labTestDocument(t, client, patientID, userID)}
	start, results := make(chan struct{}), make(chan error, 2)
	for _, id := range documents {
		go func() {
			<-start
			results <- repo.AttachDocument(ctx, report.ID, patientID, id)
		}()
	}
	close(start)
	succeeded := 0
	for range documents {
		err := <-results
		if err == nil {
			succeeded++
		} else if !errors.Is(err, labs.ErrDocumentLinkConflict) {
			t.Fatal(err)
		}
	}
	if succeeded != 1 {
		t.Fatalf("expected exactly one link, got %d", succeeded)
	}
	stored, err := repo.FindByID(ctx, report.ID)
	if err != nil || stored == nil || stored.ExamDocumentID == nil {
		t.Fatalf("missing linked report: %v", err)
	}
	if err := repo.AttachDocument(ctx, report.ID, patientID, *stored.ExamDocumentID); err != nil {
		t.Fatalf("same-document retry should succeed: %v", err)
	}
}

func TestLabsCreateDuplicatePreservesOriginal(t *testing.T) {
	for _, scenario := range []string{"fingerprint", "document"} {
		t.Run(scenario, func(t *testing.T) {
			client, repo, patientID, userID := labTestDatabase(t)
			documentID := labTestDocument(t, client, patientID, userID)
			original := labTestReport(patientID, userID)
			original.ExamDocumentID = &documentID
			fingerprint := "fixture-fingerprint"
			original.Fingerprint = &fingerprint
			ctx := context.Background()
			if err := repo.Create(ctx, original); err != nil {
				t.Fatal(err)
			}
			duplicate := labTestReport(patientID, userID)
			if scenario == "fingerprint" {
				duplicate.Fingerprint = &fingerprint
			} else {
				duplicate.ExamDocumentID = &documentID
			}
			if err := repo.Create(ctx, duplicate); !errors.Is(err, labs.ErrLabReportAlreadyExists) {
				t.Fatalf("expected duplicate error, got %v", err)
			}
			stored, err := repo.FindByID(ctx, duplicate.ID)
			if err != nil || stored != nil {
				t.Fatalf("duplicate should not persist: %v", err)
			}
			stored, err = repo.FindByID(ctx, original.ID)
			if err != nil || stored == nil || stored.ExamDocumentID == nil || *stored.ExamDocumentID != documentID || len(stored.TestResults[0].Items) != 1 {
				t.Fatalf("original report was changed: %v", err)
			}
		})
	}
}

func TestLabsRejectCrossPatientDocument(t *testing.T) {
	client, repo, patientID, userID := labTestDatabase(t)
	ctx := context.Background()
	otherPatient := uuid.New()
	if _, err := client.Pool().Exec(ctx, "INSERT INTO patients VALUES ($1)", otherPatient); err != nil {
		t.Fatal(err)
	}
	documentID := labTestDocument(t, client, otherPatient, userID)
	report := labTestReport(patientID, userID)
	report.ExamDocumentID = &documentID
	if err := repo.Create(ctx, report); !errors.Is(err, labs.ErrDocumentLinkConflict) {
		t.Fatalf("cross-patient create accepted: %v", err)
	}
	report.ExamDocumentID = nil
	if err := repo.Create(ctx, report); err != nil {
		t.Fatal(err)
	}
	if err := repo.AttachDocument(ctx, report.ID, patientID, documentID); !errors.Is(err, labs.ErrDocumentLinkConflict) {
		t.Fatalf("cross-patient attachment accepted: %v", err)
	}
}
