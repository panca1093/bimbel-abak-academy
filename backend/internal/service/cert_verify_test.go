package service

import (
	"bytes"
	"testing"
	"time"
)

func TestGenerateCertificatePDF_Classic(t *testing.T) {
	pdf, err := generateCertificatePDF("classic", "Test", "Test Exam", time.Now())
	if err != nil {
		t.Fatalf("generateCertificatePDF failed: %v", err)
	}

	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) {
		t.Errorf("expected PDF to start with %%PDF-1.4, got %q", pdf[:20])
	}

	if !bytes.HasSuffix(pdf, []byte("%%EOF\n")) {
		t.Errorf("expected PDF to end with %%EOF")
	}

	if !bytes.Contains(pdf, []byte("(Test)")) {
		t.Errorf("expected PDF to contain (Test)")
	}

	if !bytes.Contains(pdf, []byte("(Test Exam)")) {
		t.Errorf("expected PDF to contain (Test Exam)")
	}
}
