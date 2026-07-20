package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"

	"akademi-bimbel/internal/model"
)

func baseCardRegistration() *model.RegistrationDetail {
	sched := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC) // 17:00 WIB
	detail := &model.RegistrationDetail{}
	detail.ExamRegistration = model.ExamRegistration{
		ID:     uuid.New(),
		Token:  "AB12CD34",
		Status: "registered",
	}
	detail.Exam.Title = "Ujian Simulasi UTBK Saintek"
	detail.Exam.ScheduledAt = &sched
	detail.Exam.RequiresCheckin = true
	mins := 30
	detail.Exam.CheckInWindowMinutes = &mins
	return detail
}

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

// cardRasterDPI is the resolution renderToPNG rasterizes at (see
// pdftest_helper_test.go's "-r 150" pdftoppm flag).
const cardRasterDPI = 150.0

func mmToPx(mm float64) float64 { return mm * cardRasterDPI / 25.4 }

// assertSinglePageA6Landscape rasterizes the PDF (renderToPNG already fails
// on anything but exactly one page) and verifies its rendered pixel
// dimensions match 148x105mm landscape at cardRasterDPI (FR-20). Asserting
// on the rasterized page rather than the PDF's declared MediaBox catches
// content-level bugs — e.g. a full Y-axis inversion — that a MediaBox check
// alone would miss, since the page object itself stays correctly sized even
// when what's drawn on it is flipped.
func assertSinglePageA6Landscape(t *testing.T, pdfBytes []byte) {
	t.Helper()
	img := renderToPNG(t, pdfBytes)
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= h {
		t.Fatalf("expected landscape orientation, got %dx%d", w, h)
	}
	wantW, wantH := mmToPx(cardPageW), mmToPx(cardPageH)
	if diff := float64(w) - wantW; diff < -6 || diff > 6 {
		t.Errorf("page width %dpx, want ~%.0fpx (%.0fmm @%.0fdpi)", w, wantW, cardPageW, cardRasterDPI)
	}
	if diff := float64(h) - wantH; diff < -6 || diff > 6 {
		t.Errorf("page height %dpx, want ~%.0fpx (%.0fmm @%.0fdpi)", h, wantH, cardPageH, cardRasterDPI)
	}
}

func TestGenerateExamCardPDF_PhotoNull_SinglePageA6(t *testing.T) {
	detail := baseCardRegistration()
	pdf, err := generateExamCardPDF(detail, "Saifullah Panca", "Akademi Bimbel", nil, nil)
	if err != nil {
		t.Fatalf("generateExamCardPDF: %v", err)
	}
	assertSinglePageA6Landscape(t, pdf)
}

// TestGenerateExamCardPDF_HeaderBandRendersAtTop rasterizes the card and
// checks the navy header band's ink is where it belongs — at the top edge,
// not the bottom — and that the region just below it is plain background.
// A full Y-axis inversion of the card would put the band at the bottom and
// pass a page-count/orientation-only check; this catches that class of bug
// on the rendered pixels (FR-30), the same class of assertion FR-31 requires
// the certificate suite to make.
func TestGenerateExamCardPDF_HeaderBandRendersAtTop(t *testing.T) {
	detail := baseCardRegistration()
	pdf, err := generateExamCardPDF(detail, "Saifullah Panca", "Akademi Bimbel", nil, nil)
	if err != nil {
		t.Fatalf("generateExamCardPDF: %v", err)
	}
	img := renderToPNG(t, pdf)

	navyR, navyG, navyB := hexRGB(cardNavyHex)
	headerR, headerG, headerB := avgColorAt(img, cardPageW, cardPageH, 120, 8)
	if colorDistance(headerR, headerG, headerB, float64(navyR), float64(navyG), float64(navyB)) > 30*30*3 {
		t.Errorf("header band at (120mm,8mm): got (%.0f,%.0f,%.0f), want navy (%d,%d,%d) — header band may be missing or displaced", headerR, headerG, headerB, navyR, navyG, navyB)
	}

	bgR, bgG, bgB := avgColorAt(img, cardPageW, cardPageH, 120, 20)
	if colorDistance(bgR, bgG, bgB, 255, 255, 255) > 30*30*3 {
		t.Errorf("background at (120mm,20mm): got (%.0f,%.0f,%.0f), want white — header band may have leaked past its 16mm height or the card is vertically inverted", bgR, bgG, bgB)
	}
}

func TestGenerateExamCardPDF_PhotoPresent_SinglePageA6(t *testing.T) {
	detail := baseCardRegistration()
	pdf, err := generateExamCardPDF(detail, "Saifullah Panca", "Akademi Bimbel", nil, fakePNG(t))
	if err != nil {
		t.Fatalf("generateExamCardPDF: %v", err)
	}
	assertSinglePageA6Landscape(t, pdf)
}

func TestGenerateExamCardPDF_LongStudentName_SinglePageA6(t *testing.T) {
	detail := baseCardRegistration()
	longName := "Zulfikar Nurhadiningrat Wicaksono Śarma Al-Farisi bin Abdurrahman Setiawan"
	pdf, err := generateExamCardPDF(detail, longName, "Akademi Bimbel", nil, nil)
	if err != nil {
		t.Fatalf("generateExamCardPDF: %v", err)
	}
	assertSinglePageA6Landscape(t, pdf)
}

func TestGenerateExamCardPDF_LongExamTitle_SinglePageA6(t *testing.T) {
	detail := baseCardRegistration()
	detail.Exam.Title = "Ujian Simulasi Tes Potensi Skolastik dan Penalaran Umum UTBK-SNBT Gelombang Kedua Tahun Ajaran 2026/2027 untuk Seluruh Jurusan Saintek dan Soshum"
	pdf, err := generateExamCardPDF(detail, "Saifullah Panca", "Akademi Bimbel", nil, nil)
	if err != nil {
		t.Fatalf("generateExamCardPDF: %v", err)
	}
	assertSinglePageA6Landscape(t, pdf)
}

// TestGenerateExamCardPDF_MissingOrCorruptAssets_NeverFails covers FR-21: a
// missing logo/photo (nil) or corrupt bytes (unfetchable-equivalent) must
// never fail card generation.
func TestGenerateExamCardPDF_MissingOrCorruptAssets_NeverFails(t *testing.T) {
	cases := []struct {
		name  string
		logo  []byte
		photo []byte
	}{
		{"both nil", nil, nil},
		{"corrupt logo", []byte("not an image"), nil},
		{"corrupt photo", nil, []byte("not an image")},
		{"corrupt both", []byte("garbage"), []byte("garbage")},
		{"empty byte slices", []byte{}, []byte{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			detail := baseCardRegistration()
			pdf, err := generateExamCardPDF(detail, "Saifullah Panca", "Akademi Bimbel", c.logo, c.photo)
			if err != nil {
				t.Fatalf("generateExamCardPDF must not fail for missing/corrupt assets: %v", err)
			}
			assertSinglePageA6Landscape(t, pdf)
		})
	}
}

// TestGenerateExamCardPDF_TokenRendersComplete verifies FR-22 end to end via
// pdftotext: the token string must appear verbatim, never truncated, even
// under the long-name/long-title layout pressure that squeezes the rest of
// the card. Skips if poppler's pdftotext isn't available on the host.
func TestGenerateExamCardPDF_TokenRendersComplete(t *testing.T) {
	pdftotext, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Skip("pdftotext not available, skipping end-to-end token check")
	}

	cases := []struct {
		name      string
		token     string
		studentNm string
		examTitle string
	}{
		{"normal", "AB12CD34", "Saifullah Panca", "Ujian Simulasi UTBK Saintek"},
		{"long token under long title", "AB12CD34XYZ9988QRST", "Saifullah Panca", "Ujian Simulasi Tes Potensi Skolastik dan Penalaran Umum UTBK-SNBT Gelombang Kedua Tahun Ajaran 2026/2027 untuk Seluruh Jurusan Saintek dan Soshum"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			detail := baseCardRegistration()
			detail.Token = c.token
			detail.Exam.Title = c.examTitle
			pdfBytes, err := generateExamCardPDF(detail, c.studentNm, "Akademi Bimbel", nil, nil)
			if err != nil {
				t.Fatalf("generateExamCardPDF: %v", err)
			}

			tmpPDF, err := os.CreateTemp(t.TempDir(), "card-*.pdf")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tmpPDF.Write(pdfBytes); err != nil {
				t.Fatal(err)
			}
			tmpPDF.Close()

			out, err := exec.Command(pdftotext, tmpPDF.Name(), "-").Output()
			if err != nil {
				t.Fatalf("pdftotext: %v", err)
			}
			if !strings.Contains(string(out), c.token) {
				t.Errorf("expected extracted text to contain complete token %q, got:\n%s", c.token, out)
			}
		})
	}
}

func TestCardScheduleText_FormatsAsiaJakarta(t *testing.T) {
	detail := baseCardRegistration()
	got := cardScheduleText(detail)
	want := "01 Aug 2026 17:00 WIB"
	if got != want {
		t.Errorf("cardScheduleText() = %q, want %q", got, want)
	}
}

func TestCardScheduleText_NilScheduledAt(t *testing.T) {
	detail := baseCardRegistration()
	detail.Exam.ScheduledAt = nil
	if got := cardScheduleText(detail); got != "-" {
		t.Errorf("cardScheduleText() with nil ScheduledAt = %q, want %q", got, "-")
	}
}

func TestRegisterOptionalImage_NilBytesReturnsFalse(t *testing.T) {
	pdf := newCardTestPDF(t)
	if registerOptionalImage(pdf, "x", nil) {
		t.Error("expected false for nil image bytes")
	}
	if !pdf.Ok() {
		t.Error("pdf must remain in an ok state")
	}
}

func TestRegisterOptionalImage_CorruptBytesReturnsFalseAndClearsError(t *testing.T) {
	pdf := newCardTestPDF(t)
	if registerOptionalImage(pdf, "x", []byte("definitely not an image")) {
		t.Error("expected false for corrupt image bytes")
	}
	if !pdf.Ok() {
		t.Error("registerOptionalImage must clear any internal pdf error on failure")
	}
}

func TestRegisterOptionalImage_ValidPNGReturnsTrue(t *testing.T) {
	pdf := newCardTestPDF(t)
	if !registerOptionalImage(pdf, "x", fakePNG(t)) {
		t.Error("expected true for a valid PNG")
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
