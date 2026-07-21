package service

import (
	"testing"
)

// TestRenderCertificate_FieldDraggedToLowerLeft_LandsInLowerLeftNotMirrored was
// the Task 11 PDF-renderer substitute for the browser E2E check (the
// renderCertificate/renderCertificateWithImages PDF-rendering path it
// exercised was removed in Task 6 — certificates now render via
// buildCertificateHTML + Gotenberg). The HTML-pipeline equivalent of this R1
// upside-down-certificate regression guard is Task 9's job.
func TestRenderCertificate_FieldDraggedToLowerLeft_LandsInLowerLeftNotMirrored(t *testing.T) {
	t.Skip("rewritten as HTML tests in Task 9")
}
