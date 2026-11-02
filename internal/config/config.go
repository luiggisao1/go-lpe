package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	Environment        string
	SlackSigningSecret string
	RedisAddr          string
	GoogleApiKey       string
	SlackBotToken      string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:               getEnv("PORT", "8080"),
		Environment:        getEnv("ENV", "development"),
		SlackSigningSecret: getEnv("SLACK_SIGNING_SECRET", "test_signing_secret"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		GoogleApiKey:       getEnv("GOOGLE_API_KEY", "test_api_key"),
		SlackBotToken:      getEnv("SLACK_BOT_TOKEN", "test_bot_token"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
