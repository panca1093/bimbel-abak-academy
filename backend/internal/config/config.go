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
