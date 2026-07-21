//go:build gotenberg_integration

// Real-Gotenberg integration test (FR-6 acceptance gate). Excluded from the
// default suite by the gotenberg_integration build tag, so plain
// `go test ./...` never touches it. Run it against a live Gotenberg with:
//
//	GOTENBERG_URL=http://localhost:3001 \
//	  go test -tags gotenberg_integration -run TestCertificateRender_RealGotenberg ./internal/service/
//
// Set CERT_SAMPLE_OUT=/path/to/sample.pdf to also dump the rendered PDF for the
// visual acceptance check (page count, orientation, fonts, background, field
// positions) per pdf-layout-needs-visual-verification.
package service

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestCertificateRender_RealGotenberg(t *testing.T) {
	url := os.Getenv("GOTENBERG_URL")
	if url == "" {
		t.Skip("GOTENBERG_URL not set — skipping real-Gotenberg integration test")
	}

	const tmpl = "classic"
	layout := defaultLayout(tmpl)
	vals := certificateFieldValues(
		"Ujian Nasional Matematika 2026",
		"Saifullah Panca Dwiputra",
		time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC).Format("2 January 2006"),
		"ABK/2026/0001/000042",
	)
	bg := builtinCertificateBackground(tmpl)

	html, err := buildCertificateHTML(layout, vals, bg, nil)
	if err != nil {
		t.Fatalf("buildCertificateHTML: %v", err)
	}

	renderer := newGotenbergRenderer(url, &http.Client{Timeout: 30 * time.Second})
	pdf, err := renderer.RenderHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("RenderHTML against %s: %v", url, err)
	}

	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("output is not a PDF (no %%PDF- header); first bytes: %q", pdf[:min(16, len(pdf))])
	}
	if !bytes.Contains(pdf, []byte("%%EOF")) {
		t.Fatalf("PDF missing %%%%EOF trailer — likely truncated (%d bytes)", len(pdf))
	}
	if len(pdf) < 1024 {
		t.Fatalf("PDF suspiciously small (%d bytes) — expected a real rendered certificate", len(pdf))
	}

	if out := os.Getenv("CERT_SAMPLE_OUT"); out != "" {
		if err := os.WriteFile(out, pdf, 0o644); err != nil {
			t.Fatalf("write sample PDF to %s: %v", out, err)
		}
		t.Logf("wrote %d-byte sample certificate to %s for visual inspection", len(pdf), out)
	}
}
