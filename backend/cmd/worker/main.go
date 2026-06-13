package main

import (
	"akademi-bimbel/config"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"akademi-bimbel/internal/platform"
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

	w := worker.New(pool, cfg.WorkerPollInterval)
	logger.Info("worker started", "poll_interval", cfg.WorkerPollInterval.String())
	w.Run(ctx)
	logger.Info("worker stopped")
}
