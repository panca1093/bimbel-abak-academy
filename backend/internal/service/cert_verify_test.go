package service

import (
	"bytes"
	"testing"
	"time"
)

func TestGenerateCertificatePDF_Classic(t *testing.T) {
	pdf, err := generateCertificatePDF("classic", "Test", "Test Exam", time.Now(), nil)
	if err != nil {
		t.Fatalf("generateCertificatePDF failed: %v", err)
	}

	// Verify valid PDF header (gofpdf uses %PDF-1.3 by default)
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Errorf("expected PDF to start with %%PDF-, got %q", pdf[:20])
	}

	if !bytes.HasSuffix(pdf, []byte("%%EOF\n")) {
		t.Errorf("expected PDF to end with %%%%EOF")
	}

	// Compression is disabled for these generators specifically so the rendered
	// student name and exam title survive as literal, greppable substrings here —
	// this is what actually proves the per-call data reached the output, not just
	// "a PDF of plausible size came out."
	if !bytes.Contains(pdf, []byte("(Test)")) {
		t.Errorf("expected PDF to contain (Test)")
	}

	if !bytes.Contains(pdf, []byte("(Test Exam)")) {
		t.Errorf("expected PDF to contain (Test Exam)")
	}
}
