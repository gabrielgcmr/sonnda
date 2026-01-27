//Para GCS especificamente, **NÃO é necessário** criar um wrapper do `storage.Client` do Google porque:
//1. ✅ O SDK do Google já é bem abstraído
//2. ✅ Não vou trocar implementações internas do GCS
//3. ✅ Seria over-engineering

package filestorage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/domain/ports"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCSObjectStorage struct {
	client     *storage.Client
	bucketName string
	projectID  string
}

var _ ports.FileStorageService = (*GCSObjectStorage)(nil)

func NewGCSObjectStorage(
	ctx context.Context,
	bucketName,
	projectID string,
	opts ...option.ClientOption) (*GCSObjectStorage, error) {
	var (
		client *storage.Client
		err    error
	)

	if len(opts) > 0 {
		client, err = storage.NewClient(ctx, opts...)
	} else {
		client, err = storage.NewClient(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create gcs client: %w", err)
	}

	return &GCSObjectStorage{
		client:     client,
		bucketName: bucketName,
		projectID:  projectID,
	}, nil
}

func (a *GCSObjectStorage) Upload(
	ctx context.Context,
	file io.Reader,
	objectName string,
	contentType string) (string, error) {

	bucket := a.client.Bucket(a.bucketName)
	object := bucket.Object(objectName)
	writer := object.NewWriter(ctx)

	if contentType != "" {
		writer.ContentType = contentType
	}

	if _, err := io.Copy(writer, file); err != nil {
		_ = writer.Close()
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// URL pública assinada (24h)
	gcsURI := fmt.Sprintf("gs://%s/%s", a.bucketName, objectName)
	return gcsURI, nil
}

func (a *GCSObjectStorage) Delete(ctx context.Context, uri string) error {
	objectName := extractObjectName(uri, a.bucketName)

	object := a.client.Bucket(a.bucketName).Object(objectName)
	if err := object.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", uri, err)
	}
	return nil
}

func (a *GCSObjectStorage) GetSignedURL(
	ctx context.Context,
	uri string,
	expirationMinutes int,
) (string, error) {
	objectName := extractObjectName(uri, a.bucketName)

	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(time.Duration(expirationMinutes) * time.Minute),
	}

	url, err := a.client.Bucket(a.bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// Close libera o client subjacente do GCS.
func (a *GCSObjectStorage) Close() error {
	if a.client != nil {
		return a.client.Close()
	}
	return nil
}

// extractObjectName extrai o nome do objeto da URI
// Ex: "gs://bucket-name/path/to/file.pdf" -> "path/to/file.pdf"
func extractObjectName(uri, bucketName string) string {
	prefix := fmt.Sprintf("gs://%s/", bucketName)
	if len(uri) > len(prefix) && uri[:len(prefix)] == prefix {
		return uri[len(prefix):]
	}
	// Se não tiver o prefixo, assume que já é o object name
	return uri
}
