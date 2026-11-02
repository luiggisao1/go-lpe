package config

import (
	"os"
)

type Config struct {
	Port               string
	Environment        string
	SlackSigningSecret string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		Environment:        getEnv("ENV", "development"),
		SlackSigningSecret: getEnv("SLACK_SIGNING_SECRET", "test_signing_secret"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
