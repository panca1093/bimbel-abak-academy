package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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

var pageTypeRe = regexp.MustCompile(`/Type\s*/Page\b`)
var mediaBoxRe = regexp.MustCompile(`/MediaBox\s*\[\s*([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s*\]`)

// assertSinglePageA6Landscape parses the raw (uncompressed page-object)
// bytes of the PDF to verify exactly one page whose MediaBox is 148x105mm
// landscape (FR-20), without depending on external tools.
func assertSinglePageA6Landscape(t *testing.T, pdfBytes []byte) {
	t.Helper()
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Fatalf("expected PDF magic prefix, got %q", pdfBytes[:min(5, len(pdfBytes))])
	}

	pages := pageTypeRe.FindAll(pdfBytes, -1)
	if len(pages) != 1 {
		t.Fatalf("expected exactly 1 page object, found %d", len(pages))
	}

	m := mediaBoxRe.FindSubmatch(pdfBytes)
	if m == nil {
		t.Fatalf("MediaBox not found in PDF bytes")
	}
	var box [4]float64
	for i := 0; i < 4; i++ {
		v, err := strconv.ParseFloat(string(m[i+1]), 64)
		if err != nil {
			t.Fatalf("parsing MediaBox value %q: %v", m[i+1], err)
		}
		box[i] = v
	}
	wPt := box[2] - box[0]
	hPt := box[3] - box[1]
	const ptPerMM = 72.0 / 25.4
	wMM := wPt / ptPerMM
	hMM := hPt / ptPerMM

	if wMM < hMM {
		t.Fatalf("expected landscape orientation (w>h), got %.2fx%.2fmm", wMM, hMM)
	}
	if diff := wMM - 148; diff < -0.5 || diff > 0.5 {
		t.Errorf("expected page width ~148mm, got %.2fmm", wMM)
	}
	if diff := hMM - 105; diff < -0.5 || diff > 0.5 {
		t.Errorf("expected page height ~105mm, got %.2fmm", hMM)
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
