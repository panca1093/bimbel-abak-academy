package adapter

import (
	"context"
	"testing"
	"time"

	"akademi-bimbel/internal/service"
)

func TestNoopPaymentClient_CreatePayment(t *testing.T) {
	client := &NoopPaymentClient{}
	req := service.PaymentRequest{
		OrderID:   "order-123",
		Reference: "ref-123",
		Amount:    100000,
		ExpiresIn: 24 * time.Hour,
	}

	resp, err := client.CreatePayment(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.GatewayRef == "" {
		t.Error("expected non-empty GatewayRef")
	}

	if resp.PaymentURL == "" {
		t.Error("expected non-empty PaymentURL")
	}

	if resp.ExpiresAt.IsZero() {
		t.Error("expected non-zero ExpiresAt")
	}

	if resp.ExpiresAt.Before(time.Now()) {
		t.Error("expected ExpiresAt to be in the future")
	}
}

func TestNoopPaymentClient_QueryStatus(t *testing.T) {
	client := &NoopPaymentClient{}

	status, err := client.QueryStatus(context.Background(), "noop-order-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if status.Paid {
		t.Error("expected Paid=false for noop client")
	}
}

func TestNoopPaymentClient_VerifySignature(t *testing.T) {
	client := &NoopPaymentClient{}

	// No gateway is configured, so nothing could have signed this payload:
	// every signature — including the empty string — must be rejected, or an
	// unconfigured deploy leaves the unauthenticated webhook open to anyone.
	if client.VerifySignature([]byte("payload"), "signature") {
		t.Error("expected VerifySignature to return false when no gateway is configured")
	}
	if client.VerifySignature([]byte("payload"), "") {
		t.Error("expected VerifySignature to reject an empty signature")
	}
}
