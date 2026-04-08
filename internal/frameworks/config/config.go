package config

import "os"

type Config struct {
	GRPCAddress string
	HTTPAddress string
	PostgresDSN string
}

func Load() Config {
	return Config{
		GRPCAddress: envOrDefault("RESOURCE_MANAGER_GRPC_ADDR", ":8081"),
		HTTPAddress: envOrDefault("RESOURCE_MANAGER_HTTP_ADDR", ":8080"),
		PostgresDSN: envOrDefault("RESOURCE_MANAGER_POSTGRES_DSN", "postgres://resource-manager:resource-manager@localhost:5432/resource_manager?sslmode=disable"),
	}
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
