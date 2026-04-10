package config

import (
	"os"
	"strings"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Port            string
	DatabaseURL     string
	SuperadminDBURL string
	CORSOrigins     []string
	Environment     string
	Kubeconfig      string
	RedisURL        string
}

func New() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		SuperadminDBURL: getEnv("SUPERADMIN_DATABASE_URL", ""),
		CORSOrigins:     strings.Split(getEnv("CORS_ORIGINS", "*"), ","),
		Environment:     getEnv("ENVIRONMENT", "development"),
		Kubeconfig:      getEnv("KUBECONFIG", ""),
		RedisURL:        getEnv("REDIS_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
