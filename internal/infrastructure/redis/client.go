// internal/adapters/outbound/storage/data/redis/client.go
package redisstore

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient(redisURL string) (*redis.Client, error) {
	if redisURL == "" {
		return nil, errors.New("redis url is required")
	}
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	opt.MinIdleConns = 0
	opt.PoolTimeout = 2 * time.Second
	opt.ReadTimeout = 2 * time.Second
	opt.WriteTimeout = 2 * time.Second

	c := redis.NewClient(opt)

	// Ping rápido pra falhar cedo
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return c, nil
}
