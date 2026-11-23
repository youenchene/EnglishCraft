package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Address       string
	DatabaseURL   string
	CookieSecure  bool
	ShutdownGrace time.Duration
}

func Load() (Config, error) {
	config := Config{
		Address:       envOrDefault("APP_ADDRESS", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		CookieSecure:  envOrDefault("COOKIE_SECURE", "true") != "false",
		ShutdownGrace: 10 * time.Second,
	}
	if config.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return config, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
