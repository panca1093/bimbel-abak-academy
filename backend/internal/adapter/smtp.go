package adapter

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string // full "Name <addr>" override; falls back below when empty
	FromName string // display name paired with Username when From is empty
}

// SMTPProvider sends OTP codes and emails over SMTP (STARTTLS on :587). It
// satisfies both service.OTPProvider and service.EmailProvider, so a single
// instance can back both. Delivery is email-only — phone/WhatsApp OTP still
// needs a real gateway (Fazpass).
type SMTPProvider struct {
	addr string
	auth smtp.Auth
	// envelopeFrom is the bare address for the SMTP MAIL FROM — must match the
	// authenticated user (Gmail rejects otherwise). headerFrom is the From:
	// header, which may carry a display name.
	envelopeFrom string
	headerFrom   string
}

func NewSMTPProvider(cfg SMTPConfig) *SMTPProvider {
	headerFrom := cfg.From
	if headerFrom == "" {
		if cfg.FromName != "" {
			headerFrom = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.Username)
		} else {
			headerFrom = cfg.Username
		}
	}
	return &SMTPProvider{
		addr:         net.JoinHostPort(cfg.Host, cfg.Port),
		auth:         smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host),
		envelopeFrom: cfg.Username,
		headerFrom:   headerFrom,
	}
}

func (p *SMTPProvider) SendOTP(_ context.Context, channel, to, code string) error {
	if channel != "email" {
		return fmt.Errorf("smtp provider: unsupported OTP channel %q (email only)", channel)
	}
	body := fmt.Sprintf("Kode OTP kamu: %s\n\nBerlaku singkat, jangan bagikan ke siapa pun.", code)
	return p.send(to, "Kode OTP Abak Academy", body)
}

func (p *SMTPProvider) SendEmail(_ context.Context, to, subject, body string) error {
	return p.send(to, subject, body)
}

func (p *SMTPProvider) send(to, subject, body string) error {
	return smtp.SendMail(p.addr, p.auth, p.envelopeFrom, []string{to}, buildMessage(p.headerFrom, to, subject, body))
}

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}
