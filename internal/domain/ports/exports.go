// internal/domain/ports/exports.go
package ports

import (
	"sonnda-api/internal/domain/ports/ai"
	"sonnda-api/internal/domain/ports/auth"
	"sonnda-api/internal/domain/ports/data"
	"sonnda-api/internal/domain/ports/file"
)

type (
	// ai - DTOs
	ExtractedTestItem   = ai.ExtractedTestItem
	ExtractedTestResult = ai.ExtractedTestResult
	ExtractedLabReport  = ai.ExtractedLabReport

	// ai - Interfaces
	DocumentExtractorService = ai.DocumentExtractorService

	// auth
	IdentityService = auth.IdentityService

	// data - Repositories
	LabsRepo          = data.LabsRepo
	MedicalRecordRepo = data.MedicalRecordRepo
	PatientRepo       = data.PatientRepo
	PatientAccessRepo = data.PatientAccessRepo
	RequestRepo       = data.RequestRepo
	ProfessionalRepo  = data.ProfessionalRepo
	UserRepo          = data.UserRepo
	// data - DTOs
	AccessiblePatient = data.AccessiblePatient

	// file
	FileStorageService = file.FileStorageService
)
