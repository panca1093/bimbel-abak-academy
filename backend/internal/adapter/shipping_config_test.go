package adapter

import (
	"context"
	"testing"

	"akademi-bimbel/config"
	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/service"
)

// mockConfigReader implements configReader for testing.
type mockConfigReader struct {
	systemConfigRows []repository.SystemConfigRow
	listErr          error
}

func (m *mockConfigReader) ListSystemConfig(ctx context.Context) ([]repository.SystemConfigRow, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.systemConfigRows, nil
}

// TestResolveBiteshipAPIKey_EnvOnlyWhenDBEmpty tests that env value is used when DB is empty.
func TestResolveBiteshipAPIKey_EnvOnlyWhenDBEmpty(t *testing.T) {
	hexKey := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
	envKeyValue := "env-biteship-key"

	repo := &mockConfigReader{
		systemConfigRows: []repository.SystemConfigRow{
			{
				Key:      "biteship_api_key",
				Value:    "",
				IsSecret: true,
			},
		},
	}

	cfg := &config.Config{
		BiteshipAPIKey:      envKeyValue,
		ConfigEncryptionKey: hexKey,
	}

	apiKey, source := resolveBiteshipAPIKey(context.Background(), repo, cfg)

	if apiKey != envKeyValue {
		t.Errorf("expected apiKey=%q, got %q", envKeyValue, apiKey)
	}
	if source != "env" {
		t.Errorf("expected source=env, got %s", source)
	}
}

// TestResolveBiteshipAPIKey_DBWinsOverEnv tests that DB value is attempted before env.
// When DB has a non-empty encrypted value, we attempt to decrypt it.
// If decryption fails, we fall back to env. If it succeeds, source="db".
func TestResolveBiteshipAPIKey_DBWinsOverEnv(t *testing.T) {
	hexKey := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
	envKeyValue := "env-biteship-key"

	repo := &mockConfigReader{
		systemConfigRows: []repository.SystemConfigRow{
			{
				Key:      "biteship_api_key",
				Value:    "malformed-encrypted-value", // Malformed, will fail to decrypt
				IsSecret: true,
			},
		},
	}

	cfg := &config.Config{
		BiteshipAPIKey:      envKeyValue,
		ConfigEncryptionKey: hexKey,
	}

	apiKey, source := resolveBiteshipAPIKey(context.Background(), repo, cfg)

	// Since the encrypted value is malformed, decryption will fail and we'll fall back to env
	if apiKey != envKeyValue {
		t.Errorf("expected apiKey=%q (env fallback after bad decrypt), got %q", envKeyValue, apiKey)
	}
	// Source should be "env" because decryption failed
	if source != "env" {
		t.Errorf("expected source=env (after failed decrypt), got %s", source)
	}
}

// TestResolveBiteshipAPIKey_DBPrecedenceLogicWithValidEncryption tests that the DB value wins
// and source="db" when the stored ciphertext decrypts successfully, even though an env key is
// also configured.
func TestResolveBiteshipAPIKey_DBPrecedenceLogicWithValidEncryption(t *testing.T) {
	hexKey := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
	envKeyValue := "env-biteship-key"
	dbKeyValue := "db-biteship-key"

	ciphertext, err := service.EncryptConfigValue(hexKey, dbKeyValue)
	if err != nil {
		t.Fatalf("failed to encrypt fixture: %v", err)
	}

	repo := &mockConfigReader{
		systemConfigRows: []repository.SystemConfigRow{
			{
				Key:      "biteship_api_key",
				Value:    ciphertext,
				IsSecret: true,
			},
		},
	}

	cfg := &config.Config{
		BiteshipAPIKey:      envKeyValue,
		ConfigEncryptionKey: hexKey,
	}

	apiKey, source := resolveBiteshipAPIKey(context.Background(), repo, cfg)

	if apiKey != dbKeyValue {
		t.Errorf("expected apiKey=%q (DB value), got %q", dbKeyValue, apiKey)
	}
	if source != "db" {
		t.Errorf("expected source=db, got %s", source)
	}
}

// TestResolveBiteshipAPIKey_NoConfigEncryptionKey tests fallback when ConfigEncryptionKey is empty.
func TestResolveBiteshipAPIKey_NoConfigEncryptionKey(t *testing.T) {
	envKeyValue := "env-biteship-key"

	repo := &mockConfigReader{
		systemConfigRows: []repository.SystemConfigRow{
			{
				Key:      "biteship_api_key",
				Value:    "some-encrypted-value",
				IsSecret: true,
			},
		},
	}

	cfg := &config.Config{
		BiteshipAPIKey:      envKeyValue,
		ConfigEncryptionKey: "", // Empty key
	}

	apiKey, source := resolveBiteshipAPIKey(context.Background(), repo, cfg)

	if apiKey != envKeyValue {
		t.Errorf("expected apiKey=%q (env fallback), got %q", envKeyValue, apiKey)
	}
	if source != "env" {
		t.Errorf("expected source=env (fallback), got %s", source)
	}
}

// TestResolveBiteshipAPIKey_NilRepo tests fallback when repo is nil.
func TestResolveBiteshipAPIKey_NilRepo(t *testing.T) {
	envKeyValue := "env-biteship-key"

	cfg := &config.Config{
		BiteshipAPIKey:      envKeyValue,
		ConfigEncryptionKey: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
	}

	apiKey, source := resolveBiteshipAPIKey(context.Background(), nil, cfg)

	if apiKey != envKeyValue {
		t.Errorf("expected apiKey=%q (env fallback), got %q", envKeyValue, apiKey)
	}
	if source != "env" {
		t.Errorf("expected source=env (fallback), got %s", source)
	}
}

// TestResolveBiteshipAPIKey_NeitherDBNorEnv tests that apiKey is empty when neither is set.
// In this case, source is still "env" (the initial value), but ResolveShippingClient will
// log "noop" and use NoopLogisticsClient because apiKey is empty.
func TestResolveBiteshipAPIKey_NeitherDBNorEnv(t *testing.T) {
	hexKey := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"

	repo := &mockConfigReader{
		systemConfigRows: []repository.SystemConfigRow{
			{
				Key:      "biteship_api_key",
				Value:    "",
				IsSecret: true,
			},
		},
	}

	cfg := &config.Config{
		BiteshipAPIKey:      "",
		ConfigEncryptionKey: hexKey,
	}

	apiKey, source := resolveBiteshipAPIKey(context.Background(), repo, cfg)

	if apiKey != "" {
		t.Errorf("expected empty apiKey, got %q", apiKey)
	}
	if source != "env" {
		t.Errorf("expected source=env (empty but fallback), got %s", source)
	}
}

// TestResolveShippingClient_EnvKeyOnly tests that BiteshipClient is returned when env key is set.
func TestResolveShippingClient_EnvKeyOnly(t *testing.T) {
	hexKey := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"

	repo := &mockConfigReader{
		systemConfigRows: []repository.SystemConfigRow{
			{
				Key:      "biteship_api_key",
				Value:    "",
				IsSecret: true,
			},
		},
	}

	cfg := &config.Config{
		BiteshipAPIKey:      "env-biteship-key",
		ConfigEncryptionKey: hexKey,
	}

	client := ResolveShippingClient(context.Background(), repo, cfg)

	// Should return BiteshipClient, not NoopLogisticsClient
	if _, ok := client.(*BiteshipClient); !ok {
		t.Errorf("expected *BiteshipClient, got %T", client)
	}
}

// TestResolveShippingClient_NoKey tests that NoopLogisticsClient is returned when no key is set.
func TestResolveShippingClient_NoKey(t *testing.T) {
	hexKey := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"

	repo := &mockConfigReader{
		systemConfigRows: []repository.SystemConfigRow{
			{
				Key:      "biteship_api_key",
				Value:    "",
				IsSecret: true,
			},
		},
	}

	cfg := &config.Config{
		BiteshipAPIKey:      "",
		ConfigEncryptionKey: hexKey,
	}

	client := ResolveShippingClient(context.Background(), repo, cfg)

	// Should return NoopLogisticsClient
	if _, ok := client.(*service.NoopLogisticsClient); !ok {
		t.Errorf("expected *service.NoopLogisticsClient, got %T", client)
	}
}
