package adapter

import (
	"context"
	"fmt"
)

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

type NoopNotif struct{}

func (NoopNotif) Send(context.Context, string, string, string) error { return nil }

type NoopStorage struct{}

func (NoopStorage) Put(context.Context, string, []byte, string) (string, error) { return "", nil }
