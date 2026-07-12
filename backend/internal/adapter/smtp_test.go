package adapter

import (
	"context"
	"strings"
	"testing"
)

func TestSMTPProvider_SendOTP_RejectsNonEmailChannel(t *testing.T) {
	p := NewSMTPProvider(SMTPConfig{Host: "smtp.gmail.com", Port: "587", Username: "a@b.com"})
	// whatsapp/SMS can't go over SMTP — must error before any network call.
	if err := p.SendOTP(context.Background(), "whatsapp", "+628123", "123456"); err == nil {
		t.Fatal("expected error for non-email channel, got nil")
	}
}

func TestNewSMTPProvider_HeaderFrom(t *testing.T) {
	// bare username when neither From nor FromName is set
	p := NewSMTPProvider(SMTPConfig{Host: "smtp.gmail.com", Port: "587", Username: "sender@gmail.com"})
	if p.headerFrom != "sender@gmail.com" {
		t.Errorf("headerFrom: got %q, want the bare username", p.headerFrom)
	}

	// FromName pairs a display name with the username
	pn := NewSMTPProvider(SMTPConfig{Host: "h", Port: "587", Username: "sender@gmail.com", FromName: "Abak Academy"})
	if pn.headerFrom != "Abak Academy <sender@gmail.com>" {
		t.Errorf("headerFrom: got %q, want display name paired with username", pn.headerFrom)
	}

	// explicit From overrides everything
	po := NewSMTPProvider(SMTPConfig{Host: "h", Port: "587", Username: "u@x.com", From: "Abak <no-reply@x.com>", FromName: "Ignored"})
	if po.headerFrom != "Abak <no-reply@x.com>" {
		t.Errorf("headerFrom: got %q, want the explicit From", po.headerFrom)
	}

	// envelope sender is always the authenticated username, never the display form
	if pn.envelopeFrom != "sender@gmail.com" {
		t.Errorf("envelopeFrom: got %q, want the bare authenticated username", pn.envelopeFrom)
	}
}

func TestBuildMessage_HasHeadersAndBody(t *testing.T) {
	msg := string(buildMessage("from@x.com", "to@y.com", "Kode OTP", "code: 123456"))
	for _, want := range []string{
		"From: from@x.com\r\n",
		"To: to@y.com\r\n",
		"Subject: Kode OTP\r\n",
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing header %q", want)
		}
	}
	// headers and body separated by a blank CRLF line
	if !strings.Contains(msg, "\r\n\r\ncode: 123456") {
		t.Errorf("body not separated from headers by a blank line:\n%s", msg)
	}
}
