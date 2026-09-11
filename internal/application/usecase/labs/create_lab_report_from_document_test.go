// internal/application/usecase/labs/create_lab_report_from_document_test.go
package labsuc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/entity/labs"
	"github.com/gabrielgcmr/sonnda/internal/domain/entity/patient"
	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
	"github.com/gabrielgcmr/sonnda/internal/domain/repository"
	"github.com/gabrielgcmr/sonnda/internal/kernel/apperr"
	"github.com/google/uuid"
)

type labPatientRepo struct{ repository.Patient }

func (labPatientRepo) FindByID(_ context.Context, id uuid.UUID) (*patient.Patient, error) {
	return &patient.Patient{ID: id}, nil
}

type labRepoStub struct {
	repository.Labs
	existing       *labs.LabReport
	saved          *labs.LabReport
	createErr      error
	attachErr      error
	attached       bool
	race           bool
	signatureCalls int
}

func (r *labRepoStub) FindBySignature(context.Context, uuid.UUID, string) (*labs.LabReport, error) {
	r.signatureCalls++
	if r.race && r.signatureCalls == 1 {
		return nil, nil
	}
	return r.existing, nil
}

func (r *labRepoStub) FindByID(context.Context, uuid.UUID) (*labs.LabReport, error) {
	return r.existing, nil
}

func (r *labRepoStub) Create(_ context.Context, report *labs.LabReport) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.saved = report
	return nil
}

func (r *labRepoStub) AttachDocument(_ context.Context, reportID, patientID, documentID uuid.UUID) error {
	r.attached = true
	if r.attachErr != nil {
		return r.attachErr
	}
	if reportID != r.existing.ID || patientID != r.existing.PatientID {
		return errors.New("unexpected link target")
	}
	r.existing.ExamDocumentID = &documentID
	return nil
}

type labExtractorStub struct{}

func (labExtractorStub) ExtractLabReport(context.Context, string, string) (*labextraction.ExtractedLabReport, error) {
	value, unit := "99", "mg/dL"
	return &labextraction.ExtractedLabReport{Tests: []labextraction.ExtractedTestResult{
		{TestName: "Glicose", Items: []labextraction.ExtractedTestItem{
			{ParameterName: "Glicose", ResultValue: &value, ResultUnit: &unit},
		}},
	}}, nil
}

func TestCreateLabReportDocumentLinks(t *testing.T) {
	for _, scenario := range []string{"new", "legacy", "same_document", "other_document", "labs_route", "concurrent_same", "concurrent_other", "attach_conflict", "save_failure"} {
		t.Run(scenario, func(t *testing.T) {
			patientID, userID, documentID, oldDocumentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
			existing := &labs.LabReport{ID: uuid.New(), PatientID: patientID, UploadedBy: userID}
			repo := &labRepoStub{existing: existing}
			input := CreateLabReportFromDocumentInput{
				PatientID: patientID, UploadedByUserID: userID, ExamDocumentID: &documentID,
				DocumentURI: "test://exam.pdf", MimeType: "application/pdf",
			}
			switch scenario {
			case "new", "save_failure":
				repo.existing = nil
				if scenario == "save_failure" {
					repo.createErr = errors.New("database unavailable")
				}
			case "same_document", "concurrent_same":
				existing.ExamDocumentID = &documentID
			case "other_document", "concurrent_other":
				existing.ExamDocumentID = &oldDocumentID
			case "labs_route":
				input.ExamDocumentID = nil
			case "attach_conflict":
				repo.attachErr = labs.ErrDocumentLinkConflict
			}
			if scenario == "concurrent_same" || scenario == "concurrent_other" {
				repo.race = true
				repo.createErr = labs.ErrLabReportAlreadyExists
			}
			uc := NewCreateLabReportFromDocument(labPatientRepo{}, repo, labExtractorStub{})
			output, err := uc.Execute(context.Background(), input)
			if scenario == "other_document" || scenario == "concurrent_other" || scenario == "attach_conflict" || scenario == "save_failure" {
				var appErr *apperr.AppError
				if output != nil || !errors.As(err, &appErr) {
					t.Fatalf("expected application error, got output=%v err=%v", output, err)
				}
				want := apperr.RESOURCE_ALREADY_EXISTS
				if scenario == "attach_conflict" {
					want = apperr.RESOURCE_CONFLICT
				} else if scenario == "save_failure" {
					want = apperr.INFRA_DATABASE_ERROR
				}
				if appErr.Kind != want {
					t.Fatalf("expected %v, got %v", want, appErr.Kind)
				}
				if existing.ExamDocumentID != nil && *existing.ExamDocumentID != oldDocumentID {
					t.Fatal("original document was replaced")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if scenario != "labs_route" && (output.ExamDocumentID == nil || *output.ExamDocumentID != documentID) {
				t.Fatalf("document link not preserved: %v", output.ExamDocumentID)
			}
			if scenario == "new" {
				if repo.saved == nil || len(repo.saved.TestResults) != 1 || len(repo.saved.TestResults[0].Items) != 1 {
					t.Fatal("structured results were not saved")
				}
			} else if output.ID != existing.ID || repo.saved != nil {
				t.Fatal("existing report was not reused")
			}
			if repo.attached != (scenario == "legacy") {
				t.Fatalf("unexpected attach: %v", repo.attached)
			}
		})
	}
}

func TestApplyConfirmedCollectionDate(t *testing.T) {
	extractedDifferentDate := time.Date(2026, time.September, 10, 8, 30, 0, 0, time.UTC)
	extractedSameDate := time.Date(2026, time.September, 16, 8, 30, 0, 0, time.UTC)
	confirmedDate := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
	report := &labs.LabReport{TestResults: []labs.LabResult{
		{CollectedAt: &extractedDifferentDate},
		{CollectedAt: &extractedSameDate},
		{},
	}}

	applyConfirmedCollectionDate(report, &confirmedDate)

	wantConfirmed := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	if got := report.TestResults[0].CollectedAt; got == nil || !got.Equal(wantConfirmed) {
		t.Fatalf("different extracted date = %v, want %v", got, wantConfirmed)
	}
	if got := report.TestResults[1].CollectedAt; got == nil || !got.Equal(extractedSameDate) {
		t.Fatalf("same extracted date should preserve time: got %v, want %v", got, extractedSameDate)
	}
	if got := report.TestResults[2].CollectedAt; got == nil || !got.Equal(wantConfirmed) {
		t.Fatalf("missing extracted date = %v, want %v", got, wantConfirmed)
	}
}
