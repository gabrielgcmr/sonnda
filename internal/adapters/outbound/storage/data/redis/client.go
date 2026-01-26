package redisstore

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient() (*redis.Client, error) {
	url := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(url)
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
