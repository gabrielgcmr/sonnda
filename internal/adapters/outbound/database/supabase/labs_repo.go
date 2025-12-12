// internal/adapters/secondary/database/supabase/labs_repository.go
package supabase

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sonnda-api/internal/core/domain"
	"sonnda-api/internal/core/ports/repositories"

	"github.com/jackc/pgx/v5"
)

type LabsRepository struct {
	client *Client
}

// Garantia em tempo de compilação de que implementa repositories.LabsRepository
var _ repositories.LabsRepository = (*LabsRepository)(nil)

func NewLabsRepository(client *Client) repositories.LabsRepository {
	return &LabsRepository{client: client}
}

/* ============================================================
   SQL CONSTANTS
   ============================================================ */

const (
	insertLabReportSQL = `
		INSERT INTO lab_reports (
			patient_id,
			patient_name,
			patient_dob,
			lab_name,
			lab_phone,
			insurance_provider,
			requesting_doctor,
			technical_manager,
			report_date,
			raw_text,
			uploaded_by_user_id
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at, updated_at, uploaded_by_user_id
	`

	insertLabTestResultSQL = `
		INSERT INTO lab_test_results (
			lab_report_id,
			test_name,
			material,
			method,
			collected_at,
			release_at
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`

	insertLabTestItemSQL = `
		INSERT INTO lab_test_items (
			lab_test_result_id,
			parameter_name,
			result_text,
			result_unit,
			reference_text
		)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id
	`

	selectLabReportByIDSQL = `
		SELECT
			id,
			patient_id,
			patient_name,
			patient_dob,
			lab_name,
			lab_phone,
			insurance_provider,
			requesting_doctor,
			technical_manager,
			report_date,
			raw_text,
			uploaded_by_user_id,
			created_at,
			updated_at
		FROM lab_reports
		WHERE id = $1
	`

	selectLabTestResultsByReportIDSQL = `
		SELECT
			id,
			test_name,
			material,
			method,
			collected_at,
			release_at
		FROM lab_test_results
		WHERE lab_report_id = $1
		ORDER BY test_name
	`

	selectLabTestItemsByResultIDSQL = `
		SELECT
			id,
			parameter_name,
			result_text,
			result_unit,
			reference_text
		FROM lab_test_items
		WHERE lab_test_result_id = $1
		ORDER BY parameter_name
	`

	selectLabReportsByPatientIDSQL = `
		SELECT
			id,
			patient_id,
			patient_name,
			lab_name,
			report_date,
			uploaded_by_user_id,
			created_at,
			updated_at
		FROM lab_reports
		WHERE patient_id = $1
		ORDER BY report_date DESC NULLS LAST, created_at DESC
		LIMIT $2 OFFSET $3
	`

	deleteLabReportSQL = `
		DELETE FROM lab_reports
		WHERE id = $1
	`

	selectItemsTimelineByPatientAndParamSQL = `
		SELECT
			lr.id          AS report_id,
			ltr.id         AS test_result_id,
			lti.id         AS item_id,
			lr.report_date AS report_date,
			ltr.test_name  AS test_name,
			lti.parameter_name,
			lti.result_text,
			lti.result_unit
		FROM lab_test_items lti
		JOIN lab_test_results ltr ON lti.lab_test_result_id = ltr.id
		JOIN lab_reports      lr  ON ltr.lab_report_id      = lr.id
		WHERE lr.patient_id      = $1
		  AND lti.parameter_name = $2
		ORDER BY lr.report_date DESC NULLS LAST, lr.created_at DESC
		LIMIT $3 OFFSET $4
	`
)

/* ============================================================
   NULL HELPERS
   ============================================================ */

func nullableString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func nullableTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time
	return &t
}

/* ============================================================
   CREATE
   ============================================================ */

func (r *LabsRepository) Create(ctx context.Context, report *domain.LabReport) (err error) {
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

	// 1) lab_reports
	row := tx.QueryRow(ctx, insertLabReportSQL,
		report.PatientID,
		report.PatientName,
		report.PatientDOB,
		report.LabName,
		report.LabPhone,
		report.InsuranceProvider,
		report.RequestingDoctor,
		report.TechnicalManager,
		report.ReportDate,
		report.RawText,
		report.UploadedByUserID,
	)

	var uploadedBy sql.NullString
	if err = row.Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt, &uploadedBy); err != nil {
		return err
	}
	report.UploadedByUserID = nullableString(uploadedBy)

	// 2) lab_test_results + lab_test_items
	for i := range report.TestResults {
		tr := &report.TestResults[i]

		row := tx.QueryRow(ctx, insertLabTestResultSQL,
			report.ID,
			tr.TestName,
			tr.Material,
			tr.Method,
			tr.CollectedAt,
			tr.ReleaseAt,
		)

		if err = row.Scan(&tr.ID); err != nil {
			return err
		}
		tr.LabReportID = report.ID

		for j := range tr.Items {
			item := &tr.Items[j]

			row := tx.QueryRow(ctx, insertLabTestItemSQL,
				tr.ID,
				item.ParameterName,
				item.ResultValue,
				item.ResultUnit,
				item.ReferenceText,
			)

			if err = row.Scan(&item.ID); err != nil {
				return err
			}
			item.LabTestResultID = tr.ID
		}
	}

	return nil
}

/* ============================================================
   FIND BY ID (laudo completo)
   ============================================================ */

// FindByID busca um laudo completo, incluindo testes e itens.
func (r *LabsRepository) FindByID(ctx context.Context, reportID string) (*domain.LabReport, error) {
	var lr domain.LabReport

	// 1) lab_reports
	var (
		patientName, labName, labPhone, insuranceProvider,
		requestingDoctor, technicalManager, rawText, uploadedBy sql.NullString
		patientDOB, reportDate, updatedAt sql.NullTime
	)

	row := r.client.Pool().QueryRow(ctx, selectLabReportByIDSQL, reportID)

	if err := row.Scan(
		&lr.ID,
		&lr.PatientID,
		&patientName,
		&patientDOB,
		&labName,
		&labPhone,
		&insuranceProvider,
		&requestingDoctor,
		&technicalManager,
		&reportDate,
		&rawText,
		&uploadedBy,
		&lr.CreatedAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	lr.PatientName = nullableString(patientName)
	lr.PatientDOB = nullableTime(patientDOB)
	lr.LabName = nullableString(labName)
	lr.LabPhone = nullableString(labPhone)
	lr.InsuranceProvider = nullableString(insuranceProvider)
	lr.RequestingDoctor = nullableString(requestingDoctor)
	lr.TechnicalManager = nullableString(technicalManager)
	lr.ReportDate = nullableTime(reportDate)
	lr.RawText = nullableString(rawText)
	lr.UploadedByUserID = nullableString(uploadedBy)
	if t := nullableTime(updatedAt); t != nil {
		lr.UpdatedAt = *t
	}

	// 2) lab_test_results
	rows, err := r.client.Pool().Query(ctx, selectLabTestResultsByReportIDSQL, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testResults []domain.LabTestResult

	for rows.Next() {
		var (
			tr                     domain.LabTestResult
			material, method       sql.NullString
			collectedAt, releaseAt sql.NullTime
		)

		if err := rows.Scan(
			&tr.ID,
			&tr.TestName,
			&material,
			&method,
			&collectedAt,
			&releaseAt,
		); err != nil {
			return nil, err
		}

		tr.LabReportID = lr.ID
		tr.Material = nullableString(material)
		tr.Method = nullableString(method)
		tr.CollectedAt = nullableTime(collectedAt)
		tr.ReleaseAt = nullableTime(releaseAt)

		testResults = append(testResults, tr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 3) lab_test_items para cada test_result
	for i := range testResults {
		tr := &testResults[i]

		itemRows, err := r.client.Pool().Query(ctx, selectLabTestItemsByResultIDSQL, tr.ID)
		if err != nil {
			return nil, err
		}

		var items []domain.LabTestItem

		for itemRows.Next() {
			var (
				item                    domain.LabTestItem
				resultValue, resultUnit sql.NullString
				referenceText           sql.NullString
			)

			if err := itemRows.Scan(
				&item.ID,
				&item.ParameterName,
				&resultValue,
				&resultUnit,
				&referenceText,
			); err != nil {
				itemRows.Close()
				return nil, err
			}

			item.LabTestResultID = tr.ID
			item.ResultValue = nullableString(resultValue)
			item.ResultUnit = nullableString(resultUnit)
			item.ReferenceText = nullableString(referenceText)

			items = append(items, item)
		}

		itemRows.Close()

		if err := itemRows.Err(); err != nil {
			return nil, err
		}

		tr.Items = items
	}

	lr.TestResults = testResults

	return &lr, nil
}

/* ============================================================
   LIST BY PATIENT (cabeçalhos)
   ============================================================ */

// FindByPatientID retorna apenas os cabeçalhos dos laudos do paciente.
func (r *LabsRepository) FindByPatientID(ctx context.Context, patientID string, limit, offset int) ([]domain.LabReport, error) {
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

	rows, err := r.client.Pool().Query(ctx, selectLabReportsByPatientIDSQL, patientID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.LabReport

	for rows.Next() {
		var (
			lr                   domain.LabReport
			patientName, labName sql.NullString
			reportDate           sql.NullTime
			uploadedBy           sql.NullString
			updatedAt            sql.NullTime
		)

		if err := rows.Scan(
			&lr.ID,
			&lr.PatientID,
			&patientName,
			&labName,
			&reportDate,
			&uploadedBy,
			&lr.CreatedAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		lr.PatientName = nullableString(patientName)
		lr.LabName = nullableString(labName)
		lr.ReportDate = nullableTime(reportDate)
		lr.UploadedByUserID = nullableString(uploadedBy)
		if t := nullableTime(updatedAt); t != nil {
			lr.UpdatedAt = *t
		}

		result = append(result, lr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

/* ============================================================
   DELETE
   ============================================================ */

func (r *LabsRepository) Delete(ctx context.Context, reportID string) error {
	_, err := r.client.Pool().Exec(ctx, deleteLabReportSQL, reportID)
	return err
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
) ([]domain.LabTestItemTimeline, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.client.Pool().Query(ctx, selectItemsTimelineByPatientAndParamSQL,
		patientID, parameterName, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.LabTestItemTimeline

	for rows.Next() {
		var (
			item                    domain.LabTestItemTimeline
			reportDate              sql.NullTime
			resultValue, resultUnit sql.NullString
		)

		if err := rows.Scan(
			&item.ReportID,
			&item.TestResultID,
			&item.ItemID,
			&reportDate,
			&item.TestName,
			&item.ParameterName,
			&resultValue,
			&resultUnit,
		); err != nil {
			return nil, err
		}

		item.ReportDate = nullableTime(reportDate)
		item.ResultValue = nullableString(resultValue)
		item.ResultUnit = nullableString(resultUnit)

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
