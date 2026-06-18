package config

import (
	"os"
	"time"
)

type Config struct {
	Env                string
	HTTPPort           string
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	WorkerPollInterval time.Duration
	CORSOrigins        []string

	JWTSecret           string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	OTPSecret           string
	OTPTTL              time.Duration
	GoogleClientID      string
	FazpassMerchantKey  string
	FazpassAPIKey       string
	FazpassBaseURL      string
	MidtransServerKey   string
	MidtransClientKey   string
	MidtransEnv         string
}

func Load() Config {
	return Config{
		Env:                env("APP_ENV", "development"),
		HTTPPort:           env("HTTP_PORT", "8080"),
		DatabaseURL:        env("DATABASE_URL", "postgres://akademi:akademi@localhost:5432/akademi?sslmode=disable"),
		RedisAddr:          env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      env("REDIS_PASSWORD", ""),
		WorkerPollInterval: envDuration("WORKER_POLL_INTERVAL", 5*time.Second),
		CORSOrigins:        []string{env("WEB_ORIGIN", "http://localhost:3000")},

		JWTSecret:          env("JWT_SECRET", ""),
		AccessTokenTTL:     envDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:    envDuration("REFRESH_TOKEN_TTL", 168*time.Hour),
		OTPSecret:          env("OTP_SECRET", ""),
		OTPTTL:             envDuration("OTP_TTL", 5*time.Minute),
		GoogleClientID:     env("GOOGLE_CLIENT_ID", ""),
		FazpassMerchantKey: env("FAZPASS_MERCHANT_KEY", ""),
		FazpassAPIKey:      env("FAZPASS_API_KEY", ""),
		FazpassBaseURL:     env("FAZPASS_BASE_URL", "https://api.fazpass.com"),
		MidtransServerKey:  env("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey:  env("MIDTRANS_CLIENT_KEY", ""),
		MidtransEnv:        env("MIDTRANS_ENV", "sandbox"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
