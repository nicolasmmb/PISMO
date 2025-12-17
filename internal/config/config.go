package config

import "os"

type Config struct {
	Port         string
	DBDriver     string
	DBDSN        string
	LogLevel     string
	OTLPEndpoint string
}

func Load() Config {
	return Config{
		Port:         getEnv("PORT", "8080"),
		DBDriver:     getEnv("DB_DRIVER", "postgres"),
		DBDSN:        getEnv("DB_DSN", "postgres://pismo:pismo@localhost:5432/pismo?sslmode=disable"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
