// internal/domain/ports/exports.go
package ports

import (
	"github.com/gabrielgcmr/sonnda/internal/domain/ports/ai"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports/auth"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports/storage/data"
	"github.com/gabrielgcmr/sonnda/internal/domain/ports/storage/file"
)

type (
	// ai - DTOs
	ExtractedTestItem   = ai.ExtractedTestItem
	ExtractedTestResult = ai.ExtractedTestResult
	ExtractedLabReport  = ai.ExtractedLabReport

	// ai - Interfaces
	DocumentExtractorService = ai.ExtractorProvider

	// auth
	IdentityService = auth.IdentityProvider

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
	FileStorageService = file.FileStorageProvider
)
