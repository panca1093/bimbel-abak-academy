package platform

import (
	"context"
	"testing"
	"time"
)

func TestNoopPaymentClient_CreatePayment(t *testing.T) {
	client := &NoopPaymentClient{}
	req := PaymentRequest{
		OrderID:   "order-123",
		Reference: "ref-123",
		Amount:    100000,
		ExpiresIn: 24 * time.Hour,
	}

	resp, err := client.CreatePayment(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.PaymentRef == "" {
		t.Error("expected non-empty PaymentRef")
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

	result := client.VerifySignature([]byte("payload"), "signature")

	if !result {
		t.Error("expected VerifySignature to return true")
	}
}
