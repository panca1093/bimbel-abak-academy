package service

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// certVisualInspectionDir collects rasterized certificate PNGs for the
// NFR-1 manual visual-inspection gate: page orientation, field centering,
// and edge overflow are confirmed by looking at these images, not by the
// automated assertions below (memory: pdf-layout-needs-visual-verification).
const certVisualInspectionDir = "/private/tmp/claude-502/-Users-Panca-Documents-MyBook-Project-akademi-bimbel/5ca055a3-e8e4-4f30-82bf-8f4a23539dfd/scratchpad/cert-visual-check"

// rasterizeCertificatePDF writes pdfBytes to disk and rasterizes page 1 at
// 150dpi with pdftoppm, decoding the result as an image.Image. It also
// asserts exactly one page was produced (t.Fatal on a second page) — the
// direct regression guard for R1's historical "blank first page" /
// multi-page bug, per FR-30/FR-31.
func rasterizeCertificatePDF(t *testing.T, pdfBytes []byte, name string) image.Image {
	t.Helper()
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skip("pdftoppm not installed; cannot verify rasterized output")
	}

	dir := t.TempDir()
	pdfPath := filepath.Join(dir, name+".pdf")
	if err := os.WriteFile(pdfPath, pdfBytes, 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	prefix := filepath.Join(dir, name)
	cmd := exec.Command("pdftoppm", "-r", "150", "-png", pdfPath, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pdftoppm: %v\n%s", err, out)
	}

	page1 := prefix + "-1.png"
	if _, err := os.Stat(page1); err != nil {
		t.Fatalf("expected rasterized page 1: %v", err)
	}
	if _, err := os.Stat(prefix + "-2.png"); err == nil {
		t.Fatalf("%s: expected exactly one page, found a second page", name)
	}

	data, err := os.ReadFile(page1)
	if err != nil {
		t.Fatalf("read rasterized png: %v", err)
	}

	if err := os.MkdirAll(certVisualInspectionDir, 0o755); err != nil {
		t.Fatalf("mkdir visual-inspection dir: %v", err)
	}
	persisted := filepath.Join(certVisualInspectionDir, name+".png")
	if err := os.WriteFile(persisted, data, 0o644); err != nil {
		t.Fatalf("persist png for visual inspection: %v", err)
	}
	t.Logf("rasterized for visual inspection: %s", persisted)

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode rasterized png: %v", err)
	}
	return img
}

// assertA4LandscapeAspect checks the rasterized page is wider than tall and
// close to the 297:210 A4 ratio (FR-6).
func assertA4LandscapeAspect(t *testing.T, img image.Image, name string) {
	t.Helper()
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= h {
		t.Errorf("%s: expected landscape orientation, got %dx%d", name, w, h)
	}
	gotAspect := float64(w) / float64(h)
	wantAspect := float64(certificatePageWidthMm) / float64(certificatePageHeightMm)
	if diff := gotAspect - wantAspect; diff < -0.02 || diff > 0.02 {
		t.Errorf("%s: aspect ratio %.4f, want ~%.4f (A4 landscape)", name, gotAspect, wantAspect)
	}
}

// avgColorAt samples a small block centered on (xMm,yMm) and returns its
// average RGB, converting from the mm coordinate space to the rasterized
// pixel space at the image's own resolution.
func avgColorAt(img image.Image, xMm, yMm float64) (r, g, b float64) {
	bounds := img.Bounds()
	pxPerMmX := float64(bounds.Dx()) / certificatePageWidthMm
	pxPerMmY := float64(bounds.Dy()) / certificatePageHeightMm
	cx := int(xMm * pxPerMmX)
	cy := int(yMm * pxPerMmY)

	const half = 4
	var sr, sg, sb, n float64
	for dy := -half; dy <= half; dy++ {
		for dx := -half; dx <= half; dx++ {
			x, y := cx+dx, cy+dy
			if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
				continue
			}
			cr, cg, cb, _ := img.At(x, y).RGBA()
			sr += float64(cr >> 8)
			sg += float64(cg >> 8)
			sb += float64(cb >> 8)
			n++
		}
	}
	if n == 0 {
		return 0, 0, 0
	}
	return sr / n, sg / n, sb / n
}

func colorDistance(r1, g1, b1, r2, g2, b2 float64) float64 {
	dr, dg, db := r1-r2, g1-g2, b1-b2
	return dr*dr + dg*dg + db*db
}

// TestGenerateCertificatePDF_BuiltinsRenderOnePageWithBackground rasterizes
// each built-in template's output and confirms: exactly one page, A4
// landscape aspect, and — the direct regression guard for R1's "blank first
// page" bug — that a corner known to carry no stamped field still shows the
// template's own background color rather than a blank/white page.
func TestGenerateCertificatePDF_BuiltinsRenderOnePageWithBackground(t *testing.T) {
	templates := []string{"classic", "modern", "elegant"}
	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)

	for _, tmpl := range templates {
		tmpl := tmpl
		t.Run(tmpl, func(t *testing.T) {
			pdfBytes, err := generateCertificatePDF(tmpl, "Budi Santoso", "Ujian Matematika Dasar", submittedAt)
			if err != nil {
				t.Fatalf("generateCertificatePDF: %v", err)
			}
			if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
				t.Fatalf("output does not start with %%PDF magic bytes")
			}

			img := rasterizeCertificatePDF(t, pdfBytes, "builtin-"+tmpl)
			assertA4LandscapeAspect(t, img, tmpl)

			// (5mm, 5mm) is outside every field box in all three default
			// layouts (fields start at x=48.5mm; the logo field starts at
			// x=138.5mm) — background-only corner.
			rendered := img
			srcImg, err := png.Decode(bytes.NewReader(builtinCertificateBackground(defaultLayout(tmpl).Background.Ref)))
			if err != nil {
				t.Fatalf("decode source background: %v", err)
			}
			rr, rg, rb := avgColorAt(rendered, 5, 5)
			sr, sg, sb := avgColorAt(srcImg, 5, 5)
			if colorDistance(rr, rg, rb, sr, sg, sb) > 30*30*3 {
				t.Errorf("%s: rendered corner color (%.0f,%.0f,%.0f) does not match source background (%.0f,%.0f,%.0f) — background may be missing or misplaced", tmpl, rr, rg, rb, sr, sg, sb)
			}
		})
	}
}

// TestGenerateCertificatePDF_LongAndNonASCIINames renders a ~60-character
// name and a non-ASCII name (FR-7) on the classic template, confirming the
// output still rasterizes to a single A4-landscape page. Centering and
// page-edge overflow are confirmed by the required manual visual check of
// the persisted PNGs, not by pixel assertions here (NFR-1).
func TestGenerateCertificatePDF_LongAndNonASCIINames(t *testing.T) {
	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name        string
		studentName string
	}{
		{"long-name", "Muhammad Alexander Christopher Wijayakusuma Prabowo Setiawan"},
		{"non-ascii-name", "Zulfikar Nurhadi Śarma"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pdfBytes, err := generateCertificatePDF("classic", tc.studentName, "Ujian Bahasa Indonesia", submittedAt)
			if err != nil {
				t.Fatalf("generateCertificatePDF: %v", err)
			}
			img := rasterizeCertificatePDF(t, pdfBytes, tc.name)
			assertA4LandscapeAspect(t, img, tc.name)
		})
	}
}
