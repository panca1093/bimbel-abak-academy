package main

import (
	"akademi-bimbel/config"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"

	"akademi-bimbel/internal/infra"
	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/service"
	"akademi-bimbel/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infra.NewPool(ctx, cfg.DatabaseURL)
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

	storageClient := newStorageClient(cfg)
	svc := service.NewWithStore(repo, repo, rdb, nil, &service.NoopOTPProvider{}, &service.NoopEmailProvider{}, nil, nil, storageClient, &cfg)
	objectStore := worker.NewMinioObjectStore(storageClient)

	w := worker.New(pool, rdb, repo, cfg.WorkerPollInterval, sweeperInterval, repo, objectStore, svc, cfg.WorkerPollInterval, cfg.MinioPrivateBucketName)
	logger.Info("worker started", "poll_interval", cfg.WorkerPollInterval.String(), "sweeper_interval", sweeperInterval.String())
	w.Run(ctx)
	logger.Info("worker stopped")
}

func newStorageClient(cfg config.Config) *minio.Client {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		slog.Default().Error("init minio client", "err", err)
		return nil
	}
	return client
}
