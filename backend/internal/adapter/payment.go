package adapter

import (
	"context"
	"time"

	"akademi-bimbel/internal/service"
)

type NoopPaymentClient struct{}

func (n *NoopPaymentClient) CreatePayment(ctx context.Context, req service.PaymentRequest) (service.PaymentResponse, error) {
	return service.PaymentResponse{
		PaymentRef: "noop-" + req.OrderID,
		PaymentURL: "https://noop.payment/pay/" + req.OrderID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}, nil
}

func (n *NoopPaymentClient) QueryStatus(ctx context.Context, reference string) (service.PaymentStatus, error) {
	return service.PaymentStatus{Reference: reference, Paid: false}, nil
}

func (n *NoopPaymentClient) VerifySignature(payload []byte, signature string) bool {
	return true
}
