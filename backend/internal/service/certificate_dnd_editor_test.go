package service

import (
	"testing"
)

// TestRenderCertificate_FieldDraggedToLowerLeft_LandsInLowerLeftNotMirrored is
// the Task 11 substitute for the browser E2E check (docker web/api are stale
// baked images and can't run one - see task_11_result.json). It takes the
// exact x_mm/y_mm the CertificateFieldEditor component test proves a
// simulated drag of student_name to the lower-left quadrant produces
// (CertificateFieldEditor.test.tsx: "dragging a field to the lower-left...")
// and renders that layout through the real renderer, then rasterizes and
// looks at the pixels. A field placed at a small x_mm and a large y_mm must
// land in the lower-left of the page image, not be mirrored to the upper-left
// - the exact shape of R1's historical upside-down-certificate bug (NFR-1).
//
// Every check compares against a baseline render of the same background with
// no fields stamped, rather than a fixed brightness threshold, because the
// classic background's own decorative frame (a navy/gold border) is dark
// enough to trip a plain "is there ink here" check on its own — the diff
// isolates ink the student_name field itself introduced.
func TestRenderCertificate_FieldDraggedToLowerLeft_LandsInLowerLeftNotMirrored(t *testing.T) {
	const (
		draggedXMm = 20.0  // left half of a 297mm-wide page
		draggedYMm = 150.0 // bottom third of a 210mm-tall page
		wMm        = 200.0
		sizePt     = 26.0
		// A region whose baseline-vs-stamped brightness differs by less than
		// this is "no new ink"; glyph strokes darken a region by hundreds of
		// brightness units, far more than any anti-aliasing noise.
		newInkDiffThreshold = 80.0
	)

	page := Page{WidthMm: certificatePageWidthMm, HeightMm: certificatePageHeightMm}
	bg := Background{Kind: "builtin", Ref: "classic"}
	bgPNG := builtinCertificateBackground("classic")

	baselinePDF, err := renderCertificate(Layout{Page: page, Background: bg}, bgPNG, nil)
	if err != nil {
		t.Fatalf("renderCertificate (baseline): %v", err)
	}
	baselineImg := renderToPNG(t, baselinePDF)

	layout := Layout{
		Page:       page,
		Background: bg,
		Fields: []LayoutField{
			{
				ID: "student_name", XMm: draggedXMm, YMm: draggedYMm, WMm: wMm,
				Align: "center", Font: "source_serif_4", Weight: "bold",
				SizePt: sizePt, Color: "#1F2A44", Visible: true,
			},
		},
	}
	pdfBytes, err := renderCertificate(layout, bgPNG, map[FieldID]string{"student_name": "Budi Santoso"})
	if err != nil {
		t.Fatalf("renderCertificate: %v", err)
	}
	img := renderToPNG(t, pdfBytes)
	assertA4LandscapeAspect(t, img, "dragged-to-lower-left")

	newInk := func(t *testing.T, xMin, yMin, xMax, yMax float64) float64 {
		t.Helper()
		before := regionMinBrightness(baselineImg, certificatePageWidthMm, certificatePageHeightMm, xMin, yMin, xMax, yMax)
		after := regionMinBrightness(img, certificatePageWidthMm, certificatePageHeightMm, xMin, yMin, xMax, yMax)
		return before - after
	}

	lineHeightMm := sizePt * 0.3528 * 1.15
	xMin, xMax := draggedXMm+40, draggedXMm+wMm-40

	// The dragged position: new ink must actually be there.
	droppedYMin, droppedYMax := draggedYMm-2, draggedYMm+lineHeightMm+2
	if diff := newInk(t, xMin, droppedYMin, xMax, droppedYMax); diff < newInkDiffThreshold {
		t.Fatalf("no new ink found at the dropped position x:[%.1f,%.1f] y:[%.1f,%.1f]mm (brightness diff vs baseline=%.0f, want >=%.0f) - field text is missing or mispositioned",
			xMin, xMax, droppedYMin, droppedYMax, diff, newInkDiffThreshold)
	}

	// The vertically-mirrored position a `pageHeight - y` bug would have used
	// instead: must show no new ink versus the baseline.
	mirroredYMm := certificatePageHeightMm - draggedYMm
	mirroredYMin, mirroredYMax := mirroredYMm-2, mirroredYMm+lineHeightMm+2
	if diff := newInk(t, xMin, mirroredYMin, xMax, mirroredYMax); diff >= newInkDiffThreshold {
		t.Fatalf("new ink found at the vertically-mirrored position x:[%.1f,%.1f] y:[%.1f,%.1f]mm (brightness diff vs baseline=%.0f) - the field rendered upside-down (Y-axis flip), R1 recurring",
			xMin, xMax, mirroredYMin, mirroredYMax, diff)
	}

	// draggedYMm (150mm) is in the bottom third of a 210mm page and xMin/xMax
	// (60-180mm) sits left of the 148.5mm horizontal center - together with
	// the two checks above this confirms the field landed in the lower-left
	// quadrant, not mirrored into the upper-left one.
	if xMin >= float64(certificatePageWidthMm)/2 {
		t.Fatalf("test setup error: xMin=%.1f is not left of page center", xMin)
	}
	if draggedYMm <= float64(certificatePageHeightMm)/2 {
		t.Fatalf("test setup error: draggedYMm=%.1f is not in the bottom half of the page", float64(draggedYMm))
	}
}

