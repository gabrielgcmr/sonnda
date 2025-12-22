// internal/adapters/secondary/database/supabase/labs_repository.go
package supabase

import (
	"context"
	"errors"
	"strings"

	labssqlc "sonnda-api/internal/adapters/outbound/database/sqlc/labs"
	"sonnda-api/internal/core/domain/medicalRecord/exam/lab"
	"sonnda-api/internal/core/ports/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LabsRepository struct {
	client  *Client
	queries *labssqlc.Queries
}

// Garantia em tempo de compilação de que implementa repositories.LabsRepository
var _ repositories.LabsRepository = (*LabsRepository)(nil)

func NewLabsRepository(client *Client) repositories.LabsRepository {
	return &LabsRepository{
		client:  client,
		queries: labssqlc.New(client.Pool()),
	}
}

/* ============================================================
   CREATE
   ============================================================ */

func (r *LabsRepository) Create(ctx context.Context, report *lab.LabReport) (err error) {
	tx, err := r.client.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	qtx := labssqlc.New(tx)

	// ---------
	// 1) lab_reports
	// ---------

	if report.ID == "" {
		report.ID = uuid.NewString()
	}

	dbReport, err := qtx.CreateLabReport(ctx, labssqlc.CreateLabReportParams{
		ID:                report.ID,
		PatientID:         report.PatientID,
		PatientName:       ToText(report.PatientName),
		PatientDob:        ToDate(report.PatientDOB),
		LabName:           ToText(report.LabName),
		LabPhone:          ToText(report.LabPhone),
		InsuranceProvider: ToText(report.InsuranceProvider),
		RequestingDoctor:  ToText(report.RequestingDoctor),
		TechnicalManager:  ToText(report.TechnicalManager),
		ReportDate:        ToDate(report.ReportDate),
		RawText:           ToText(report.RawText),
		UploadedByUserID:  ToTextValue(report.UploadedByUserID),
		Fingerprint:       ToText(report.Fingerprint),
	})

	if err != nil {
		return err
	}

	// Atualiza o domain com o que voltou do banco (sem Scan!)
	report.ID = dbReport.ID
	report.CreatedAt = dbReport.CreatedAt.Time
	report.UpdatedAt = dbReport.UpdatedAt.Time
	report.UploadedByUserID = *FromText(dbReport.UploadedByUserID)
	report.Fingerprint = FromText(dbReport.Fingerprint)

	// ---------
	// 2) lab_results + lab_result_items
	// ---------

	for i := range report.TestResults {
		tr := &report.TestResults[i]

		if tr.ID == "" {
			tr.ID = uuid.NewString()
		}

		dbResID, err := qtx.CreateLabResult(ctx, labssqlc.CreateLabResultParams{
			ID:          tr.ID,
			LabReportID: report.ID,
			TestName:    tr.TestName,
			Material:    ToText(tr.Material),
			Method:      ToText(tr.Method),
			CollectedAt: ToTimestamptz(tr.CollectedAt),
			ReleaseAt:   ToTimestamptz(tr.ReleaseAt),
		})
		if err != nil {
			return err
		}

		tr.ID = dbResID
		tr.LabReportID = report.ID

		for j := range tr.Items {
			item := &tr.Items[j]

			if item.ID == "" {
				item.ID = uuid.NewString()
			}

			dbItemID, err := qtx.CreateLabResultItem(ctx, labssqlc.CreateLabResultItemParams{
				ID:            item.ID,
				LabResultID:   dbResID,
				ParameterName: item.ParameterName,
				ResultValue:   ToText(item.ResultValue),
				ResultUnit:    ToText(item.ResultUnit),
				ReferenceText: ToText(item.ReferenceText),
			})
			if err != nil {
				return err
			}

			item.ID = dbItemID
			item.LabResultID = tr.ID
		}
	}

	return nil
}

/* ============================================================
   FIND BY ID (laudo completo)
   ============================================================ */

// FindByID busca um laudo completo, incluindo testes e itens.
func (r *LabsRepository) FindByID(ctx context.Context, reportID string) (*lab.LabReport, error) {
	q := labssqlc.New(r.client.Pool())

	// 1) lab_reports
	dbReport, err := q.GetLabReportByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	lr := &lab.LabReport{
		ID:                dbReport.ID,
		PatientID:         dbReport.PatientID,
		PatientName:       FromText(dbReport.PatientName),
		PatientDOB:        FromDate(dbReport.PatientDob),
		LabName:           FromText(dbReport.LabName),
		LabPhone:          FromText(dbReport.LabPhone),
		InsuranceProvider: FromText(dbReport.InsuranceProvider),
		RequestingDoctor:  FromText(dbReport.RequestingDoctor),
		TechnicalManager:  FromText(dbReport.TechnicalManager),
		ReportDate:        FromDate(dbReport.ReportDate),
		RawText:           FromText(dbReport.RawText),
		UploadedByUserID:  *FromText(dbReport.UploadedByUserID),
		Fingerprint:       FromText(dbReport.Fingerprint),
		CreatedAt:         dbReport.CreatedAt.Time,
		UpdatedAt:         dbReport.UpdatedAt.Time,
	}

	// 2) lab_results
	dbResults, err := q.ListLabResultsByReportID(ctx, dbReport.ID)
	if err != nil {
		return nil, err
	}

	lr.TestResults = make([]lab.LabResult, 0, len(dbResults))

	for _, rrow := range dbResults {
		tr := lab.LabResult{
			ID:          rrow.ID,
			LabReportID: lr.ID,
			TestName:    rrow.TestName,
			Material:    FromText(rrow.Material),
			Method:      FromText(rrow.Method),
			CollectedAt: FromTimestamptz(rrow.CollectedAt),
			ReleaseAt:   FromTimestamptz(rrow.ReleaseAt),
		}

		// 3) items do result
		dbItems, err := q.ListLabResultItemsByResultID(ctx, rrow.ID)
		if err != nil {
			return nil, err
		}

		tr.Items = make([]lab.LabResultItem, 0, len(dbItems))
		for _, irow := range dbItems {
			item := lab.LabResultItem{
				ID:            irow.ID,
				LabResultID:   tr.ID,
				ParameterName: irow.ParameterName,
				ResultValue:   FromText(irow.ResultValue),
				ResultUnit:    FromText(irow.ResultUnit),
				ReferenceText: FromText(irow.ReferenceText),
			}
			tr.Items = append(tr.Items, item)
		}

		lr.TestResults = append(lr.TestResults, tr)
	}

	return lr, nil
}

/* ============================================================
   LIST BY PATIENT (cabeçalhos)
   ============================================================ */

// FindByPatientID retorna apenas os cabeçalhos dos laudos do paciente.
func (r *LabsRepository) FindByPatientID(ctx context.Context,
	patientID string,
	limit, offset int,
) ([]lab.LabReport, error) {
	const (
		defaultLimit  = 100
		defaultOffset = 0
	)

	if limit <= 0 {
		limit = defaultLimit
	}
	if offset < 0 {
		offset = defaultOffset
	}

	q := labssqlc.New(r.client.Pool())

	rows, err := q.ListLabReportsByPatientID(ctx, labssqlc.ListLabReportsByPatientIDParams{
		PatientID: patientID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]lab.LabReport, 0, len(rows))

	for _, row := range rows {
		createdAt, err := MustTime(row.CreatedAt)
		if err != nil {
			return nil, err
		}
		updatedAt, err := MustTime(row.UpdatedAt)
		if err != nil {
			return nil, err
		}

		lr := lab.LabReport{
			ID:               row.ID,
			PatientID:        row.PatientID,
			PatientName:      FromText(row.PatientName),
			LabName:          FromText(row.LabName),
			ReportDate:       FromDate(row.ReportDate),
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
			UploadedByUserID: *FromText(row.UploadedByUserID),
			Fingerprint:      FromText(row.Fingerprint),
		}

		out = append(out, lr)
	}

	return out, nil
}

/* ============================================================
   DELETE
   ============================================================ */

func (r *LabsRepository) Delete(ctx context.Context, reportID string) error {
	tx, err := r.client.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	qtx := labssqlc.New(tx)

	affected, err := qtx.DeleteLabReport(ctx, reportID)
	if err != nil {
		return err
	}

	if affected == 0 {
		return nil
	}

	return nil

}

/* ============================================================
   TIMELINE POR PARÂMETRO
   ============================================================ */

// ListItemsByPatientAndParameter retorna o histórico de um parâmetro específico
// (ex.: todas as creatininas) para um paciente.
func (r *LabsRepository) ListItemsByPatientAndParameter(
	ctx context.Context,
	patientID string,
	parameterName string,
	limit, offset int,
) ([]lab.LabResultItemTimeline, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	if strings.TrimSpace(parameterName) == "" {
		return nil, lab.ErrInvalidInput
	}

	q := labssqlc.New(r.client.Pool())

	rows, err := q.ListLabItemTimelineByPatientAndParameter(
		ctx,
		labssqlc.ListLabItemTimelineByPatientAndParameterParams{
			PatientID:     patientID,
			ParameterName: parameterName,
			Limit:         int32(limit),
			Offset:        int32(offset),
		},
	)
	if err != nil {
		return nil, err
	}

	out := make([]lab.LabResultItemTimeline, 0, len(rows))
	for _, row := range rows {
		item := lab.LabResultItemTimeline{
			ReportID:      row.ReportID,
			LabResultID:   row.LabResultID,
			ItemID:        row.ItemID,
			ReportDate:    FromDate(row.ReportDate), // report_date é DATE no Supabase
			TestName:      row.TestName,
			ParameterName: row.ParameterName,
			ResultValue:   FromText(row.ResultValue),
			ResultUnit:    FromText(row.ResultUnit),
		}

		out = append(out, item)
	}

	return out, nil
}

/* ============================================================
   Dedupe
   ============================================================ */

// Verifica se já existe um laudo com a mesma assinatura
func (r *LabsRepository) ExistsBySignature(
	ctx context.Context,
	patientID string,
	fingerprint string,
) (bool, error) {
	q := labssqlc.New(r.client.Pool())

	exists, err := q.ExistsLabReportByPatientAndFingerprint(ctx,
		labssqlc.ExistsLabReportByPatientAndFingerprintParams{
			PatientID:   patientID,
			Fingerprint: ToTextValue(fingerprint),
		},
	)
	if err != nil {
		return false, err
	}
	return exists, nil
}
