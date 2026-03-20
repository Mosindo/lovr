package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	StripeSecretKey     string
	StripeWebhookSecret string
	StripePriceID       string
	AppBaseURL          string
}

func Load() (Config, error) {
	cfg := Config{
		Port:                getenv("PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		StripeSecretKey:     strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")),
		StripeWebhookSecret: strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET")),
		StripePriceID:       strings.TrimSpace(os.Getenv("STRIPE_PRICE_ID")),
		AppBaseURL:          strings.TrimSpace(os.Getenv("APP_BASE_URL")),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required (for Docker/CI use postgres://postgres:postgres@postgres:5432/app?sslmode=disable)")
	}
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
