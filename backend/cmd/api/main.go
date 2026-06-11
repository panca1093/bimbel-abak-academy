package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"akademi-bimbel/internal/config"
	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/server"
	"akademi-bimbel/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()
	ctx := context.Background()

	pool, err := platform.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	rdb := platform.NewRedis(cfg.RedisAddr, cfg.RedisPassword)
	defer rdb.Close()

	repo := repository.New(pool)
	svc := service.New(repo, rdb)
	h := handler.New(svc)
	e := server.New(h, cfg)

	go func() {
		if err := e.Start(":" + cfg.HTTPPort); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server", "err", err)
			os.Exit(1)
		}
	}()
	logger.Info("api started", "port", cfg.HTTPPort, "env", cfg.Env)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "err", err)
	}
	logger.Info("api stopped")
}
