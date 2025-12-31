package labs

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewLabReport_Success(t *testing.T) {
	patientID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	uploadedBy := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	report, err := NewLabReport("  "+patientID.String()+"  ", "  "+uploadedBy.String()+"  ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if report == nil {
		t.Fatalf("expected report, got nil")
	}
	if isNilUUID(report.ID.String()) {
		t.Fatalf("expected generated ID")
	}
	if report.PatientID.String() != patientID.String() {
		t.Fatalf("expected trimmed patientID, got %q", report.PatientID)
	}
	if report.UploadedBy.String() != uploadedBy.String() {
		t.Fatalf("expected trimmed uploadedBy, got %q", report.UploadedBy)
	}
	if report.CreatedAt.Location() != time.UTC {
		t.Fatalf("expected CreatedAt in UTC")
	}
	if report.UpdatedAt.Location() != time.UTC {
		t.Fatalf("expected UpdatedAt in UTC")
	}
}

func TestNewLabReport_InvalidInputs(t *testing.T) {
	cases := []struct {
		name       string
		patientID  string
		uploadedBy string
		wantErr    error
	}{
		{name: "missing patientID", patientID: "", uploadedBy: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", wantErr: ErrInvalidPatientID},
		{name: "missing uploadedBy", patientID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", uploadedBy: "  ", wantErr: ErrInvalidUploadedByUser},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewLabReport(tc.patientID, tc.uploadedBy)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestLabReport_Normalize_TrimsOptionalAndUTC(t *testing.T) {
	report, err := NewLabReport("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	name := strPtr("  Ana  ")
	labName := strPtr("   ")
	labPhone := strPtr(" 11 9999  ")
	insurance := strPtr("  Health  ")
	requestingDoctor := strPtr(" Dr. X ")
	technicalManager := strPtr("   ")
	fingerprint := strPtr("  fp  ")
	rawText := strPtr("   ")

	dob := timePtr(time.Date(2024, 5, 10, 8, 30, 0, 0, time.FixedZone("BRT", -3*3600)))
	reportDate := timePtr(time.Date(2024, 5, 11, 9, 0, 0, 0, time.FixedZone("BRT", -3*3600)))

	report.PatientName = name
	report.LabName = labName
	report.LabPhone = labPhone
	report.InsuranceProvider = insurance
	report.RequestingDoctor = requestingDoctor
	report.TechnicalManager = technicalManager
	report.Fingerprint = fingerprint
	report.RawText = rawText
	report.PatientDOB = dob
	report.ReportDate = reportDate
	report.UpdatedAt = time.Date(2024, 5, 12, 10, 0, 0, 0, time.FixedZone("BRT", -3*3600))

	report.Normalize()

	if report.PatientName == nil || *report.PatientName != "Ana" {
		t.Fatalf("expected patient name trimmed")
	}
	if report.LabName != nil {
		t.Fatalf("expected lab name to be nil after trimming blanks")
	}
	if report.LabPhone == nil || *report.LabPhone != "11 9999" {
		t.Fatalf("expected lab phone trimmed")
	}
	if report.InsuranceProvider == nil || *report.InsuranceProvider != "Health" {
		t.Fatalf("expected insurance provider trimmed")
	}
	if report.RequestingDoctor == nil || *report.RequestingDoctor != "Dr. X" {
		t.Fatalf("expected requesting doctor trimmed")
	}
	if report.TechnicalManager != nil {
		t.Fatalf("expected technical manager to be nil after trimming blanks")
	}
	if report.Fingerprint == nil || *report.Fingerprint != "fp" {
		t.Fatalf("expected fingerprint trimmed")
	}
	if report.RawText != nil {
		t.Fatalf("expected raw text to be nil after trimming blanks")
	}
	if report.PatientDOB == nil || report.PatientDOB.Location() != time.UTC {
		t.Fatalf("expected patient DOB coerced to UTC")
	}
	if report.ReportDate == nil || report.ReportDate.Location() != time.UTC {
		t.Fatalf("expected report date coerced to UTC")
	}
	if report.UpdatedAt.Location() != time.UTC {
		t.Fatalf("expected UpdatedAt coerced to UTC")
	}
}

func TestNewLabResult_Success(t *testing.T) {
	labReportID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	result, err := NewLabResult("  "+labReportID.String()+"  ", "  Hemograma  ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatalf("expected result, got nil")
	}
	if isNilUUID(result.ID.String()) {
		t.Fatalf("expected generated ID")
	}
	if result.LabReportID.String() != labReportID.String() {
		t.Fatalf("expected trimmed labReportID, got %q", result.LabReportID)
	}
	if result.TestName != "Hemograma" {
		t.Fatalf("expected trimmed test name, got %q", result.TestName)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected items slice initialized")
	}
}

func TestNewLabResult_InvalidInputs(t *testing.T) {
	cases := []struct {
		name        string
		labReportID string
		testName    string
		wantErr     error
	}{
		{name: "missing labReportID", labReportID: "  ", testName: "x", wantErr: ErrMissingId},
		{name: "missing testName", labReportID: "cccccccc-cccc-cccc-cccc-cccccccccccc", testName: "   ", wantErr: ErrInvalidTestName},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewLabResult(tc.labReportID, tc.testName)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestLabResult_Normalize_TrimsOptionalAndUTC(t *testing.T) {
	result, err := NewLabResult("cccccccc-cccc-cccc-cccc-cccccccccccc", "Hemograma")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	material := strPtr("  sangue  ")
	method := strPtr("   ")
	collected := timePtr(time.Date(2024, 6, 1, 7, 0, 0, 0, time.FixedZone("BRT", -3*3600)))
	release := timePtr(time.Date(2024, 6, 2, 8, 0, 0, 0, time.FixedZone("BRT", -3*3600)))

	result.Material = material
	result.Method = method
	result.CollectedAt = collected
	result.ReleaseAt = release

	result.Normalize()

	if result.Material == nil || *result.Material != "sangue" {
		t.Fatalf("expected material trimmed")
	}
	if result.Method != nil {
		t.Fatalf("expected method to be nil after trimming blanks")
	}
	if result.CollectedAt == nil || result.CollectedAt.Location() != time.UTC {
		t.Fatalf("expected collectedAt coerced to UTC")
	}
	if result.ReleaseAt == nil || result.ReleaseAt.Location() != time.UTC {
		t.Fatalf("expected releaseAt coerced to UTC")
	}
}

func TestNewLabResultItem_Success(t *testing.T) {
	labResultID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	item, err := NewLabResultItem("  "+labResultID.String()+"  ", "  Hemacias  ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if item == nil {
		t.Fatalf("expected item, got nil")
	}
	if isNilUUID(item.ID.String()) {
		t.Fatalf("expected generated ID")
	}
	if item.LabResultID.String() != labResultID.String() {
		t.Fatalf("expected trimmed labResultID, got %q", item.LabResultID)
	}
	if item.ParameterName != "Hemacias" {
		t.Fatalf("expected trimmed parameter name, got %q", item.ParameterName)
	}
}

func TestNewLabResultItem_InvalidInputs(t *testing.T) {
	cases := []struct {
		name          string
		labResultID   string
		parameterName string
		wantErr       error
	}{
		{name: "missing labResultID", labResultID: "  ", parameterName: "x", wantErr: ErrMissingId},
		{name: "missing parameterName", labResultID: "dddddddd-dddd-dddd-dddd-dddddddddddd", parameterName: "   ", wantErr: ErrInvalidParameterName},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewLabResultItem(tc.labResultID, tc.parameterName)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestLabResultItem_Normalize_TrimsOptional(t *testing.T) {
	item, err := NewLabResultItem("dddddddd-dddd-dddd-dddd-dddddddddddd", "Hemacias")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	value := strPtr("  4.5  ")
	unit := strPtr("   g/dL  ")
	ref := strPtr("   ")

	item.ResultValue = value
	item.ResultUnit = unit
	item.ReferenceText = ref

	item.Normalize()

	if item.ResultValue == nil || *item.ResultValue != "4.5" {
		t.Fatalf("expected result value trimmed")
	}
	if item.ResultUnit == nil || *item.ResultUnit != "g/dL" {
		t.Fatalf("expected result unit trimmed")
	}
	if item.ReferenceText != nil {
		t.Fatalf("expected reference text nil after trimming blanks")
	}
}

func strPtr(s string) *string { return &s }

func timePtr(t time.Time) *time.Time { return &t }

func isNilUUID(raw string) bool {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return true
	}
	return parsed == uuid.Nil
}
