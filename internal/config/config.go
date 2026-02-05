// internal/config/config.go
package config

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Storage  StorageConfig
}
