package adapter

import (
	"context"
	"log/slog"

	"akademi-bimbel/config"
	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/service"
)

// ResolvePaymentClient reads Midtrans credentials from DB (system_config) first,
// falling back to env vars. Returns NoopPaymentClient only when neither source
// provides a server key.
func ResolvePaymentClient(ctx context.Context, repo *repository.Repository, cfg *config.Config) service.PaymentClient {
	serverKey, clientKey, midtransEnv, source := resolvePaymentKeys(ctx, repo, cfg)

	if serverKey != "" {
		slog.Info("payment client resolved", "source", source)
		return NewMidtransClient(serverKey, clientKey, midtransEnv)
	}
	slog.Info("payment client resolved", "source", "noop")
	return &service.NoopPaymentClient{}
}

func resolvePaymentKeys(ctx context.Context, repo *repository.Repository, cfg *config.Config) (serverKey, clientKey, midtransEnv, source string) {
	serverKey = cfg.MidtransServerKey
	clientKey = cfg.MidtransClientKey
	midtransEnv = cfg.MidtransEnv
	source = "env"

	if repo == nil || cfg.ConfigEncryptionKey == "" {
		if serverKey == "" {
			source = "none"
		}
		return
	}

	rows, err := repo.ListSystemConfig(ctx)
	if err != nil {
		slog.Warn("reading system_config for payment keys, using env fallback", "err", err)
		if serverKey == "" {
			source = "none"
		}
		return
	}

	for _, row := range rows {
		switch row.Key {
		case "midtrans_server_key":
			if row.IsSecret && row.Value != "" {
				if decrypted, decErr := service.DecryptConfigValue(cfg.ConfigEncryptionKey, row.Value); decErr == nil {
					serverKey = decrypted
					source = "db"
				}
			}
		case "midtrans_client_key":
			if row.IsSecret && row.Value != "" {
				if decrypted, decErr := service.DecryptConfigValue(cfg.ConfigEncryptionKey, row.Value); decErr == nil {
					clientKey = decrypted
				}
			}
		case "midtrans_env":
			if row.Value != "" {
				midtransEnv = row.Value
			}
		}
	}
	return
}
