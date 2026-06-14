package main

import (
	"akademi-bimbel/config"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := platform.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	defer rdb.Close()

	repo := repository.New(pool)
	sweeperInterval := 5 * time.Minute // default 5m
	w := worker.New(pool, rdb, repo, cfg.WorkerPollInterval, sweeperInterval)
	logger.Info("worker started", "poll_interval", cfg.WorkerPollInterval.String(), "sweeper_interval", sweeperInterval.String())
	w.Run(ctx)
	logger.Info("worker stopped")
}
