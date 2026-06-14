package platform

import "context"

// Pluggable third-party integrations. Vendor choice is a deploy-time decision,
// not a code decision (TRD core principle). The Noop implementations let the
// skeleton boot before any real provider is wired.

type NotifClient interface {
	Send(ctx context.Context, channel, to, message string) error
}

type StorageClient interface {
	Put(ctx context.Context, key string, body []byte, contentType string) (url string, err error)
}

type (
	NoopNotif   struct{}
	NoopStorage struct{}
)

func (NoopNotif) Send(context.Context, string, string, string) error { return nil }

func (NoopStorage) Put(context.Context, string, []byte, string) (string, error) { return "", nil }
