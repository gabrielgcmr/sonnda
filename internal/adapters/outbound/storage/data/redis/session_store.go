// internal/adapters/outbound/storage/data/redis/session_store.go
package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gabrielgcmr/sonnda/internal/domain/ports/storage/data"
	"github.com/gabrielgcmr/sonnda/internal/kernel/security"
)

const defaultSessionPrefix = "session:"

type SessionStore struct {
	client    *redis.Client
	keyPrefix string
}

var _ data.SessionStore = (*SessionStore)(nil)

func NewSessionStore(client *redis.Client, keyPrefix string) (*SessionStore, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	keyPrefix = strings.TrimSpace(keyPrefix)
	if keyPrefix == "" {
		keyPrefix = defaultSessionPrefix
	}
	return &SessionStore{
		client:    client,
		keyPrefix: keyPrefix,
	}, nil
}

type sessionPayload struct {
	Identity security.Identity `json:"identity"`
}

func (s *SessionStore) Save(ctx context.Context, sessionID string, identity security.Identity, ttl time.Duration) error {
	if strings.TrimSpace(sessionID) == "" {
		return errors.New("session id is required")
	}
	if ttl <= 0 {
		return errors.New("session ttl must be positive")
	}

	raw, err := json.Marshal(sessionPayload{Identity: identity})
	if err != nil {
		return err
	}

	return s.client.Set(ctx, s.key(sessionID), raw, ttl).Err()
}

func (s *SessionStore) Find(ctx context.Context, sessionID string) (*security.Identity, bool, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, false, nil
	}

	raw, err := s.client.Get(ctx, s.key(sessionID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var payload sessionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, false, err
	}

	return &payload.Identity, true, nil
}

func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}
	return s.client.Del(ctx, s.key(sessionID)).Err()
}

func (s *SessionStore) key(sessionID string) string {
	return s.keyPrefix + sessionID
}
