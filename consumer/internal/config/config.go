package config

import "os"

type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
	RedisAddr    string
}

func Load() Config {
	return Config{
		KafkaBrokers: splitCSV(envOrDefault("KAFKA_BROKERS", "localhost:19092")),
		KafkaTopic:   envOrDefault("KAFKA_TOPIC", "listening-events"),
		KafkaGroupID: envOrDefault("KAFKA_GROUP_ID", "feature-consumer"),
		RedisAddr:    envOrDefault("REDIS_ADDR", "localhost:6379"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
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
