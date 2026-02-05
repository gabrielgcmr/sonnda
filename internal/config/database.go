// internal/config/database.go
package config

const (
	envDatabaseURL = "DATABASE_URL"
	envRedisURL    = "REDIS_URL"
)

type DatabaseConfig struct {
	URL      string
	RedisURL string
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		URL:      getEnv(envDatabaseURL),
		RedisURL: getEnv(envRedisURL),
	}
}
