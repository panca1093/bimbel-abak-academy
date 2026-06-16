package service

import (
	"context"
	"fmt"
	"time"
)

type OTPProvider interface {
	SendOTP(ctx context.Context, channel, to, code string) error
}

type EmailProvider interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

type PaymentRequest struct {
	OrderID   string
	Reference string
	Amount    int64
	ExpiresIn time.Duration
}

type PaymentResponse struct {
	GatewayRef string
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

type ShippingQuoteRequest struct {
	DestinationZip string
	WeightGrams    int
}

type CourierRate struct {
	Courier       string
	Service       string
	EstimatedDays int
	Price         int64
}

type LogisticsClient interface {
	GetRates(ctx context.Context, req ShippingQuoteRequest) ([]CourierRate, error)
}

type NotifClient interface {
	Send(ctx context.Context, channel, to, message string) error
}

type StorageClient interface {
	Put(ctx context.Context, key string, body []byte, contentType string) (url string, err error)
}

// PreferredOTPChannel returns the best channel+destination given available contact info.
func PreferredOTPChannel(phone, email string) (channel, destination string) {
	if phone != "" {
		return "whatsapp", phone
	}
	if email != "" {
		return "email", email
	}
	return "email", ""
}

type NoopOTPProvider struct{}

func (n *NoopOTPProvider) SendOTP(_ context.Context, channel, to, code string) error {
	fmt.Printf("[noop-otp] channel=%s to=%s code=%s\n", channel, to, code)
	return nil
}

type NoopEmailProvider struct{}

func (n *NoopEmailProvider) SendEmail(_ context.Context, to, subject, body string) error {
	fmt.Printf("[noop-email] to=%s subject=%s body=%s\n", to, subject, body)
	return nil
}

type NoopPaymentClient struct{}

func (n *NoopPaymentClient) CreatePayment(ctx context.Context, req PaymentRequest) (PaymentResponse, error) {
	return PaymentResponse{
		GatewayRef: "noop-" + req.OrderID,
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

type NoopLogisticsClient struct{}

func (n *NoopLogisticsClient) GetRates(ctx context.Context, req ShippingQuoteRequest) ([]CourierRate, error) {
	return []CourierRate{
		{Courier: "JNE", Service: "REG", EstimatedDays: 3, Price: 15000},
		{Courier: "TIKI", Service: "ONS", EstimatedDays: 1, Price: 25000},
	}, nil
}
