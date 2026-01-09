package labsuc

import "github.com/google/uuid"

type CreateLabReportFromDocumentInput struct {
	PatientID        uuid.UUID
	DocumentURI      string
	MimeType         string
	UploadedByUserID uuid.UUID
}
