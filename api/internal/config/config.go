package config

import "os"

type Config struct {
	APIAddr     string
	DatabaseURL string
}

func Load() Config {
	return Config{
		APIAddr:     envOrDefault("API_ADDR", ":8080"),
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://echorec:echorec@localhost:5432/echorec?sslmode=disable"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
