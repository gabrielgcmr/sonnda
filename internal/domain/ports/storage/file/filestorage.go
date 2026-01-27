package file

import (
	"context"
	"io"
)

type FileStorageProvider interface {
	Upload(ctx context.Context, file io.Reader, objectName, contentType string) (string, error)
	Delete(ctx context.Context, uri string) error
	GetSignedURL(ctx context.Context, uri string, expirationMinutes int) (string, error)
}
