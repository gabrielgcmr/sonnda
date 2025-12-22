package lab

import "errors"

var (
	ErrLabReportNotFound      = errors.New("lab report not found")
	ErrInvalidDocument        = errors.New("invalid document")
	ErrDocumentProcessing     = errors.New("document processing failed")
	ErrInvalidDateFormat      = errors.New("invalid date format")
	ErrInvalidInput           = errors.New("invalid input")
	ErrMissingIdentifiers     = errors.New("missing identifiers")
	ErrLabReportAlreadyExists = errors.New("lab report already exists")
)
