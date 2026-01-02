package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort        string
	Environment    string
	DatabaseURL    string
	JWTSecret      string
	SupabaseURL    string
	SupabaseKey    string
	SupabaseBucket string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:        getEnv("APP_PORT", "8080"),
		Environment:    getEnv("ENV", "development"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		SupabaseURL:    os.Getenv("SUPABASE_URL"),
		SupabaseKey:    os.Getenv("SUPABASE_KEY"),
		SupabaseBucket: getEnv("SUPABASE_BUCKET", "papers"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	if cfg.SupabaseURL == "" || cfg.SupabaseKey == "" {
		return nil, errors.New("SUPABASE_URL and SUPABASE_KEY are required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
