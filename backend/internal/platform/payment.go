package platform

import (
	"context"
	"time"
)

type PaymentRequest struct {
	OrderID   string
	Reference string
	Amount    int64
	ExpiresIn time.Duration
}

type PaymentResponse struct {
	PaymentRef string
	PaymentURL string
	ExpiresAt  time.Time
}

type PaymentStatus struct {
	Reference string
	Paid      bool
	PaidAt    *time.Time
}

type PaymentClient interface {
	CreatePayment(ctx context.Context, req PaymentRequest) (PaymentResponse, error)
	QueryStatus(ctx context.Context, reference string) (PaymentStatus, error)
	VerifySignature(payload []byte, signature string) bool
}

type NoopPaymentClient struct{}

func (n *NoopPaymentClient) CreatePayment(ctx context.Context, req PaymentRequest) (PaymentResponse, error) {
	return PaymentResponse{
		PaymentRef: "noop-" + req.OrderID,
		PaymentURL: "https://noop.payment/pay/" + req.OrderID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}, nil
}

func (n *NoopPaymentClient) QueryStatus(ctx context.Context, reference string) (PaymentStatus, error) {
	return PaymentStatus{Reference: reference, Paid: false}, nil
}

func (n *NoopPaymentClient) VerifySignature(payload []byte, signature string) bool {
	return true
}
