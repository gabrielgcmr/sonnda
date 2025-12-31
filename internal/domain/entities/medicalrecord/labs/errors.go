package labs

import "errors"

var (
	ErrLabReportNotFound      = errors.New("lab report not found")
	ErrInvalidDocument        = errors.New("invalid document")
	ErrDocumentProcessing     = errors.New("document processing failed")
	ErrInvalidDateFormat      = errors.New("invalid date format")
	ErrInvalidInput           = errors.New("invalid input")
	ErrMissingId              = errors.New("missing id")
	ErrLabReportAlreadyExists = errors.New("lab report already exists")
	ErrInvalidPatientID       = errors.New("patient id is required")
	ErrInvalidUploadedByUser  = errors.New("uploaded by user id is required")
	ErrInvalidTestName        = errors.New("test name is required")
	ErrInvalidParameterName   = errors.New("parameter name is required")
)
