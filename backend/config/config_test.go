package config

import (
	"os"
	"testing"
)

func TestLoad_ConfigEncryptionKey(t *testing.T) {
	const key = "CONFIG_ENCRYPTION_KEY"
	os.Setenv(key, "test-encryption-key")
	defer os.Unsetenv(key)

	cfg := Load()
	if cfg.ConfigEncryptionKey != "test-encryption-key" {
		t.Errorf("ConfigEncryptionKey: got %q want %q", cfg.ConfigEncryptionKey, "test-encryption-key")
	}
}

func TestLoad_ConfigEncryptionKey_default(t *testing.T) {
	os.Unsetenv("CONFIG_ENCRYPTION_KEY")
	cfg := Load()
	if cfg.ConfigEncryptionKey != "" {
		t.Errorf("ConfigEncryptionKey default: got %q want %q", cfg.ConfigEncryptionKey, "")
	}
}

func TestLoad_MinioPrivateBucketName_default(t *testing.T) {
	os.Unsetenv("MINIO_PRIVATE_BUCKET_NAME")
	cfg := Load()
	if cfg.MinioPrivateBucketName != "akademi-bimbel-private" {
		t.Errorf("MinioPrivateBucketName default: got %q want %q", cfg.MinioPrivateBucketName, "akademi-bimbel-private")
	}
}
