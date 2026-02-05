// internal/config/auth.go
package config

const (
	envSupabaseProjectURL = "SUPABASE_PROJECT_URL"
	envSupabaseJWTIssuer  = "SUPABASE_JWT_ISSUER"
	envSupabaseJWTAud     = "SUPABASE_JWT_AUDIENCE"
)

type AuthConfig struct {
	SupabaseProjectURL  string
	SupabaseJWTIssuer   string
	SupabaseJWTAudience string
}

func loadAuthConfig() AuthConfig {
	return AuthConfig{
		SupabaseProjectURL:  getEnv(envSupabaseProjectURL),
		SupabaseJWTIssuer:   getEnv(envSupabaseJWTIssuer),
		SupabaseJWTAudience: getEnv(envSupabaseJWTAud),
	}
}
