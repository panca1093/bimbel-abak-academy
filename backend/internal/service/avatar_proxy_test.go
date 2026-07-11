package service

import (
	"context"
	"errors"
	"testing"

	"akademi-bimbel/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// The avatar read-proxy is unauthenticated, and avatars share one bucket with
// certificates and private student PII. OpenAvatar must therefore refuse any key
// outside the avatars/ prefix (and reject path traversal) before it ever touches
// storage — otherwise the proxy would leak certs and PII to anyone.
func TestOpenAvatar_RejectsNonAvatarKeys(t *testing.T) {
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("ak", "sk", ""),
		Secure: false,
	})
	if err != nil {
		t.Fatalf("minio.New: %v", err)
	}
	s := &Service{storage: client, cfg: &config.Config{ObjectStorageBucketName: "bucket"}}

	blocked := []string{
		"certificates/00000000-0000-0000-0000-000000000000.pdf",
		"student-bulk/school-1/import.csv",
		"avatars/../certificates/leak.pdf",
		"random-key",
		"",
	}
	for _, key := range blocked {
		if _, _, err := s.OpenAvatar(context.Background(), key); !errors.Is(err, ErrUploadNotFound) {
			t.Errorf("OpenAvatar(%q): expected ErrUploadNotFound, got %v", key, err)
		}
	}
}
