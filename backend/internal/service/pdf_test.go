package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/jung-kurt/gofpdf"
)

func fakePNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 60, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 60; x++ {
			img.Set(x, y, color.RGBA{R: 120, G: 160, B: 220, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestRegisterOptionalImage_NilBytesReturnsFalse(t *testing.T) {
	pdf := newCardTestPDF(t)
	if ok, _, _ := registerOptionalImage(pdf, "x", nil); ok {
		t.Error("expected false for nil image bytes")
	}
	if !pdf.Ok() {
		t.Error("pdf must remain in an ok state")
	}
}

func TestRegisterOptionalImage_CorruptBytesReturnsFalseAndClearsError(t *testing.T) {
	pdf := newCardTestPDF(t)
	if ok, _, _ := registerOptionalImage(pdf, "x", []byte("definitely not an image")); ok {
		t.Error("expected false for corrupt image bytes")
	}
	if !pdf.Ok() {
		t.Error("registerOptionalImage must clear any internal pdf error on failure")
	}
}

func TestRegisterOptionalImage_ValidPNGReturnsTrue(t *testing.T) {
	pdf := newCardTestPDF(t)
	ok, w, h := registerOptionalImage(pdf, "x", fakePNG(t))
	if !ok {
		t.Error("expected true for a valid PNG")
	}
	if w != 60 || h != 80 {
		t.Errorf("expected the source's own 60x80 dimensions, got %dx%d", w, h)
	}
	if !pdf.Ok() {
		t.Error("pdf must remain in an ok state after a valid registration")
	}
}

func TestShrinkToFit_NeverModifiesText_ShrinksUntilFitOrFloor(t *testing.T) {
	pdf := newCardTestPDF(t)
	token := "AB12CD34XYZ9988QRSTUVWXYZ0011"
	size := shrinkToFit(pdf, FontSourceSerif4, "B", token, 30, 20, 8)
	if size < 8 || size > 20 {
		t.Fatalf("shrinkToFit returned out-of-range size %v", size)
	}
	pdf.SetFont(FontSourceSerif4, "B", size)
	if w := pdf.GetStringWidth(token); size > 8 && w > 30 {
		t.Errorf("shrunk text still overflows maxWidth: width=%.2f maxWidth=30", w)
	}
}

func TestTruncateWithEllipsis_NeverExceedsMaxWidth(t *testing.T) {
	pdf := newCardTestPDF(t)
	pdf.SetFont(FontPublicSans, "", 9)
	text := "This is a very long piece of text that will not fit in a narrow box"
	got := truncateWithEllipsis(pdf, text, 30)
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
	if w := pdf.GetStringWidth(got); w > 30 {
		t.Errorf("truncated text still overflows: width=%.2f maxWidth=30, text=%q", w, got)
	}
}

func TestWrapLines_LongExamTitle_NoLineOverflowsAndSignalsTruncation(t *testing.T) {
	pdf := newCardTestPDF(t)
	pdf.SetFont(FontPublicSans, "", 9)
	longTitle := "Ujian Simulasi Tes Potensi Skolastik dan Penalaran Umum UTBK-SNBT Gelombang Kedua Tahun Ajaran 2026/2027 untuk Seluruh Jurusan Saintek dan Soshum"
	lines := wrapLines(pdf, longTitle, 106, 2)

	if len(lines) > 2 {
		t.Fatalf("expected at most 2 lines, got %d: %v", len(lines), lines)
	}
	for _, l := range lines {
		if w := pdf.GetStringWidth(l); w > 106 {
			t.Errorf("line %q overflows maxWidth: width=%.2f maxWidth=106", l, w)
		}
	}
	if !strings.HasSuffix(lines[len(lines)-1], "…") {
		t.Errorf("expected last line to signal truncation with an ellipsis, got %q", lines[len(lines)-1])
	}
}

func TestWrapLines_ShortText_NoTruncation(t *testing.T) {
	pdf := newCardTestPDF(t)
	pdf.SetFont(FontPublicSans, "", 9)
	lines := wrapLines(pdf, "Finals", 106, 2)
	if len(lines) != 1 || lines[0] != "Finals" {
		t.Errorf("expected single unmodified line, got %v", lines)
	}
}

func TestHexRGB(t *testing.T) {
	r, g, b := hexRGB(cardNavyHex)
	if r != 0x22 || g != 0x31 || b != 0x5B {
		t.Errorf("hexRGB(%q) = (%d,%d,%d), want (34,49,91)", cardNavyHex, r, g, b)
	}
}

func TestPtToMM(t *testing.T) {
	got := ptToMM(72)
	if diff := got - 25.4; diff < -0.001 || diff > 0.001 {
		t.Errorf("ptToMM(72) = %v, want 25.4", got)
	}
}

func newCardTestPDF(t *testing.T) *gofpdf.Fpdf {
	t.Helper()
	pdf := gofpdf.New("L", "mm", "A6", "")
	if err := RegisterFonts(pdf); err != nil {
		t.Fatalf("RegisterFonts: %v", err)
	}
	pdf.AddPage()
	return pdf
}
