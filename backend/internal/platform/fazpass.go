package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type FazpassConfig struct {
	MerchantKey string
	APIKey      string
	BaseURL     string
}

type FazpassProvider struct {
	cfg    FazpassConfig
	client *http.Client
}

func NewFazpassProvider(cfg FazpassConfig) *FazpassProvider {
	return &FazpassProvider{cfg: cfg, client: &http.Client{}}
}

func (f *FazpassProvider) SendOTP(ctx context.Context, channel, to, code string) error {
	payload := map[string]string{
		"phone_number": to,
		"otp_code":     code,
		"gateway_key":  f.cfg.APIKey,
	}
	return f.post(ctx, f.cfg.BaseURL+"/v2/otp/send-by-contact", payload)
}

// SendEmail posts to Fazpass email endpoint. If Fazpass doesn't support email in
// a given deployment, swap BaseURL to empty string and it will return an error at
// the HTTP level — no special casing needed.
func (f *FazpassProvider) SendEmail(ctx context.Context, to, subject, body string) error {
	payload := map[string]string{
		"to":      to,
		"subject": subject,
		"body":    body,
	}
	return f.post(ctx, f.cfg.BaseURL+"/v2/email/send", payload)
}

func (f *FazpassProvider) post(ctx context.Context, url string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.cfg.APIKey)
	req.Header.Set("x-merchant-key", f.cfg.MerchantKey)

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("fazpass: unexpected status %d", resp.StatusCode)
	}
	return nil
}
