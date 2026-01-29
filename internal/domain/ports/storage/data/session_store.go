// internal/domain/ports/storage/data/session_store.go
package data

import (
	"context"
	"time"

	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

// SessionStore persists session -> identity mapping with a TTL.
// It is used by web auth (cookie sessions) and can be shared by other adapters.
type SessionStore interface {
	Save(ctx context.Context, sessionID string, identity security.Identity, ttl time.Duration) error
	Find(ctx context.Context, sessionID string) (*security.Identity, bool, error)
	Delete(ctx context.Context, sessionID string) error
}
