package platform

import "context"

// Pluggable third-party integrations. Vendor choice is a deploy-time decision,
// not a code decision (TRD core principle). The Noop implementations let the
// skeleton boot before any real provider is wired.

type PaymentClient interface {
	CreateCharge(ctx context.Context, orderID string, amount int64) (snapToken string, err error)
}

type LogisticsClient interface {
	Rates(ctx context.Context, originPostal, destPostal string, weightGrams int) ([]CourierRate, error)
}

type CourierRate struct {
	Courier string
	Service string
	Cost    int64
	ETD     string
}

type NotifClient interface {
	Send(ctx context.Context, channel, to, message string) error
}

type StorageClient interface {
	Put(ctx context.Context, key string, body []byte, contentType string) (url string, err error)
}

type (
	NoopPayment   struct{}
	NoopLogistics struct{}
	NoopNotif     struct{}
	NoopStorage   struct{}
)

func (NoopPayment) CreateCharge(context.Context, string, int64) (string, error) { return "", nil }

func (NoopLogistics) Rates(context.Context, string, string, int) ([]CourierRate, error) {
	return nil, nil
}

func (NoopNotif) Send(context.Context, string, string, string) error { return nil }

func (NoopStorage) Put(context.Context, string, []byte, string) (string, error) { return "", nil }
