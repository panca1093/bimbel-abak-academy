package adapter

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"akademi-bimbel/config"
	"akademi-bimbel/internal/service"
)

// ResolveShippingClient reads Biteship API key from DB (system_config) first,
// falling back to env vars. Returns NoopLogisticsClient only when neither source
// provides an API key.
func ResolveShippingClient(ctx context.Context, repo configReader, cfg *config.Config) service.LogisticsClient {
	apiKey, source := resolveBiteshipAPIKey(ctx, repo, cfg)

	if apiKey != "" {
		slog.Info("shipping client resolved", "source", source)
		httpClient := &http.Client{Timeout: 10 * time.Second}
		return NewBiteshipClient(repo, apiKey, "https://api.biteship.com", httpClient)
	}
	slog.Info("shipping client resolved", "source", "noop")
	return &service.NoopLogisticsClient{}
}

func resolveBiteshipAPIKey(ctx context.Context, repo configReader, cfg *config.Config) (apiKey, source string) {
	apiKey = cfg.BiteshipAPIKey
	source = "env"

	if repo == nil || cfg.ConfigEncryptionKey == "" {
		if apiKey == "" {
			source = "none"
		}
		return
	}

	rows, err := repo.ListSystemConfig(ctx)
	if err != nil {
		slog.Warn("reading system_config for biteship_api_key, using env fallback", "err", err)
		if apiKey == "" {
			source = "none"
		}
		return
	}

	for _, row := range rows {
		if row.Key == "biteship_api_key" {
			if row.IsSecret && row.Value != "" {
				if decrypted, decErr := service.DecryptConfigValue(cfg.ConfigEncryptionKey, row.Value); decErr == nil {
					apiKey = decrypted
					source = "db"
					break
				}
			}
		}
	}
	return
}
