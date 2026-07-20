package service

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jung-kurt/gofpdf"
)

const fontTestString = "Zulfikar Nurhadi Śarma"

// allFontFamilies lists the FR-7a closed set in a fixed order so the
// rendered sample PDF is deterministic for visual inspection.
var allFontFamilies = []string{
	FontSourceSerif4,
	FontPublicSans,
	FontCinzel,
	FontPlayfairDisplay,
	FontCormorantGaramond,
	FontGreatVibes,
}

func TestResolveFontFamily_KnownFamiliesResolveToThemselves(t *testing.T) {
	t.Parallel()
	for _, family := range allFontFamilies {
		if got := ResolveFontFamily(family); got != family {
			t.Errorf("ResolveFontFamily(%q) = %q, want %q", family, got, family)
		}
	}
}

func TestResolveFontFamily_UnknownFallsBackToBrandDefault(t *testing.T) {
	t.Parallel()
	got := ResolveFontFamily("comic_sans")
	if got != defaultFontFamily {
		t.Errorf("ResolveFontFamily(unknown) = %q, want brand default %q", got, defaultFontFamily)
	}
}

func TestResolveFontFamily_UnknownFamilyStillRendersWithoutError(t *testing.T) {
	t.Parallel()
	pdf := gofpdf.New("P", "mm", "A4", "")
	if err := RegisterFonts(pdf); err != nil {
		t.Fatalf("RegisterFonts: %v", err)
	}
	pdf.AddPage()
	pdf.SetFont(ResolveFontFamily("this_family_does_not_exist"), "", 14)
	pdf.CellFormat(0, 10, fontTestString, "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty PDF output")
	}
}

// TestRegisterFonts_AllSixFamiliesRenderNonASCIIName renders the sample
// string in all six bundled families on one page, rasterizes it with
// pdftoppm, and writes the PNG next to the test binary so a human can
// visually confirm (a) non-ASCII glyphs render as letters, not boxes, and
// (b) the six families are visually distinct (Task 2 "Done when", NFR-1).
func TestRegisterFonts_AllSixFamiliesRenderNonASCIIName(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	if err := RegisterFonts(pdf); err != nil {
		t.Fatalf("RegisterFonts: %v", err)
	}
	pdf.AddPage()

	y := 20.0
	for _, family := range allFontFamilies {
		pdf.SetXY(15, y)
		pdf.SetFont(family, "", 18)
		pdf.CellFormat(0, 12, family+": "+fontTestString, "", 1, "L", false, 0, "")
		y += 30
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}

	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "fonts-sample.pdf")
	if err := os.WriteFile(pdfPath, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skip("pdftoppm not on PATH; skipping rasterization (visual check must be run manually)")
	}

	pngPrefix := filepath.Join(dir, "fonts-sample")
	cmd := exec.Command("pdftoppm", "-r", "150", "-png", pdfPath, pngPrefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pdftoppm: %v\n%s", err, out)
	}

	pngPath := pngPrefix + "-1.png"
	info, err := os.Stat(pngPath)
	if err != nil {
		t.Fatalf("expected rasterized PNG at %s: %v", pngPath, err)
	}
	if info.Size() == 0 {
		t.Fatal("rasterized PNG is empty")
	}

	// Persist a copy for manual visual inspection outside the test's temp dir.
	persisted := filepath.Join(os.TempDir(), "pdffonts-sample-1.png")
	data, err := os.ReadFile(pngPath)
	if err != nil {
		t.Fatalf("ReadFile rasterized PNG: %v", err)
	}
	if err := os.WriteFile(persisted, data, 0o644); err != nil {
		t.Fatalf("persist PNG for visual inspection: %v", err)
	}
	t.Logf("rasterized sample for visual inspection: %s", persisted)
}
