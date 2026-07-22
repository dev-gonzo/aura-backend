package config

import (
	"os"
)

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/editorial?sslmode=disable"

type AppConfig struct {
	AppName              string
	AppPort              string
	DatabaseURL          string
	JWTSecret            string
	InitialAdminEmail    string
	InitialAdminCPF      string
	InitialAdminPassword string
	SupabaseURL          string
	SupabaseAnonKey      string
	SupabaseS3Endpoint   string
	SupabaseS3Region     string
	SupabaseS3Bucket     string
	SupabaseS3AccessKey  string
	SupabaseS3SecretKey  string
	SupabasePublicURL    string
}

func Load() AppConfig {
	return AppConfig{
		AppName:              getenv("EDITORA_APP_NAME", "Aura Editora Backend"),
		AppPort:              getenv("EDITORA_BACKEND_PORT", "8081"),
		DatabaseURL:          getenv("DATABASE_URL", defaultDatabaseURL),
		JWTSecret:            getenv("JWT_SECRET", "aura-local-secret-change-me"),
		InitialAdminEmail:    getenv("INITIAL_ADMIN_EMAIL", "admin@aura.local"),
		InitialAdminCPF:      getenv("INITIAL_ADMIN_CPF", "52998224725"),
		InitialAdminPassword: getenv("INITIAL_ADMIN_PASSWORD", "admin"),
		SupabaseURL:          getenv("SUPABASE_URL", ""),
		SupabaseAnonKey:      getenv("SUPABASE_ANON_KEY", ""),
		SupabaseS3Endpoint:   getenv("SUPABASE_STORAGE_S3_ENDPOINT", ""),
		SupabaseS3Region:     getenv("SUPABASE_STORAGE_REGION", "us-east-1"),
		SupabaseS3Bucket:     getenv("SUPABASE_STORAGE_BUCKET", "aura-docs"),
		SupabaseS3AccessKey:  getenv("SUPABASE_STORAGE_ACCESS_KEY_ID", ""),
		SupabaseS3SecretKey:  getenv("SUPABASE_STORAGE_SECRET_ACCESS_KEY", ""),
		SupabasePublicURL:    getenv("SUPABASE_STORAGE_PUBLIC_URL", ""),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
