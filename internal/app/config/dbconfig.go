package config

import (
	"time"

	postgress "github.com/gabrielgcmr/sonnda/internal/adapters/outbound/storage/data/postgres"
)

func SupabaseConfig(cfg Config) postgress.Config {
	return postgress.Config{
		DatabaseURL:     cfg.DBURL,
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}
}
