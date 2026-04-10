package config

import (
	"os"
	"strings"
	"time"
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
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
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
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret-change-me"),
		AccessTokenTTL:  getDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getDurationEnv("REFRESH_TOKEN_TTL", 30*24*time.Hour),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}
