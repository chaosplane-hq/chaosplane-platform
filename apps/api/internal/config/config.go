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
	CSRFSecret      string
	FrontendURL     string

	GoogleClientID        string
	GoogleClientSecret    string
	GitHubClientID        string
	GitHubClientSecret    string
	MicrosoftClientID     string
	MicrosoftClientSecret string

	SESRegion    string
	SESAccessKey string
	SESSecretKey string
	SESFromEmail string

	LLMProvider string
	LLMAPIKey   string
	LLMModel    string

	StripeWebhookSecret string
	TossWebhookSecret   string
	DodoWebhookSecret   string

	Demo bool
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
		CSRFSecret:      getEnv("CSRF_SECRET", getEnv("JWT_SECRET", "dev-secret-change-me")),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),

		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:        getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret:    getEnv("GITHUB_CLIENT_SECRET", ""),
		MicrosoftClientID:     getEnv("MICROSOFT_CLIENT_ID", ""),
		MicrosoftClientSecret: getEnv("MICROSOFT_CLIENT_SECRET", ""),

		SESRegion:    getEnv("SES_REGION", "us-east-1"),
		SESAccessKey: getEnv("SES_ACCESS_KEY", ""),
		SESSecretKey: getEnv("SES_SECRET_KEY", ""),
		SESFromEmail: getEnv("SES_FROM_EMAIL", ""),

		LLMProvider: getEnv("LLM_PROVIDER", "openai"),
		LLMAPIKey:   getEnv("LLM_API_KEY", ""),
		LLMModel:    getEnv("LLM_MODEL", "gpt-4o"),

		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		TossWebhookSecret:   getEnv("TOSS_WEBHOOK_SECRET", ""),
		DodoWebhookSecret:   getEnv("DODO_WEBHOOK_SECRET", ""),

		Demo: getBoolEnv("DEMO", false),
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

func getBoolEnv(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}
