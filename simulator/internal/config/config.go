package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL string
	KafkaBrokers []string
	KafkaTopic   string
	Interval     time.Duration
	UserCount    int
}

func Load() Config {
	return Config{
		DatabaseURL:  envOrDefault("DATABASE_URL", "postgres://echorec:echorec@localhost:5432/echorec?sslmode=disable"),
		KafkaBrokers: splitCSV(envOrDefault("KAFKA_BROKERS", "localhost:19092")),
		KafkaTopic:   envOrDefault("KAFKA_TOPIC", "listening-events"),
		Interval:     envDurationOrDefault("SIMULATOR_INTERVAL", 2*time.Second),
		UserCount:    envIntOrDefault("SIMULATOR_USER_COUNT", 20),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func splitCSV(raw string) []string {
	parts := make([]string, 0, 1)
	start := 0
	for i := 0; i <= len(raw); i++ {
		if i == len(raw) || raw[i] == ',' {
			part := raw[start:i]
			if part != "" {
				parts = append(parts, part)
			}
			start = i + 1
		}
	}
	if len(parts) == 0 {
		return []string{raw}
	}
	return parts
}
