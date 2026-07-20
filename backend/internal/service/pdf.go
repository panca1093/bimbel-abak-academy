package service

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"

	"akademi-bimbel/internal/model"
)

// Exam card geometry (OQ4): A6 landscape, 148 x 105 mm, origin top-left, Y
// down — identical to gofpdf's "mm" unit mode. No Y flip anywhere.
const (
	cardPageW = 148.0
	cardPageH = 105.0

	cardNavyHex        = "#22315B"
	cardGoldHex        = "#D99A2B"
	cardGoldBgHex      = "#F8F1DD"
	cardPlaceholderHex = "#B7BECE"
)

// generateExamCardPDF renders the fixed A6-landscape exam card (FR-20..24).
// logoImg and photoImg are already-fetched image bytes (or nil) — fetching
// app_logo_url / User.PhotoURL is I/O that belongs at the call site
// (Service.GetExamCard in exam.go), so a network failure there never fails
// card generation (FR-21): a missing/unfetchable/corrupt logo just omits the
// mark, and a missing/corrupt photo draws the placeholder silhouette in the
// identical frame position.
func generateExamCardPDF(reg *model.RegistrationDetail, studentName, tenantName string, logoImg, photoImg []byte) ([]byte, error) {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "L",
		UnitStr:        "mm",
		Size:           gofpdf.SizeType{Wd: cardPageH, Ht: cardPageW},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	if err := RegisterFonts(pdf); err != nil {
		return nil, err
	}
	pdf.AddPage()

	drawCardHeaderBand(pdf, tenantName, logoImg)
	drawCardPhotoFrame(pdf, photoImg)
	drawCardDetailColumn(pdf, reg, studentName)
	drawCardTokenBand(pdf, reg.Token)
	drawCardFooterNote(pdf, reg)
	drawCardBorder(pdf)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawCardHeaderBand(pdf *gofpdf.Fpdf, tenantName string, logoImg []byte) {
	navyR, navyG, navyB := hexRGB(cardNavyHex)
	goldR, goldG, goldB := hexRGB(cardGoldHex)

	pdf.SetFillColor(navyR, navyG, navyB)
	pdf.Rect(0, 0, cardPageW, 16, "F")

	if registerOptionalImage(pdf, "card-logo", logoImg) {
		pdf.ImageOptions("card-logo", 6, 3, 10, 10, false, gofpdf.ImageOptions{}, 0, "")
	}

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(FontSourceSerif4, "B", 12)
	pdf.SetXY(20, 3)
	pdf.CellFormat(cardPageW-20-4, 7, "KARTU PESERTA UJIAN", "", 0, "L", false, 0, "")

	if tenantName != "" {
		pdf.SetTextColor(goldR, goldG, goldB)
		pdf.SetFont(FontPublicSans, "", 7)
		pdf.SetXY(20, 10)
		pdf.CellFormat(cardPageW-20-4, 5, fitOneLine(pdf, tenantName, cardPageW-20-4), "", 0, "L", false, 0, "")
	}
}

const (
	cardPhotoX = 8.0
	cardPhotoY = 24.0
	cardPhotoW = 22.0
	cardPhotoH = 28.0
)

func drawCardPhotoFrame(pdf *gofpdf.Fpdf, photoImg []byte) {
	navyR, navyG, navyB := hexRGB(cardNavyHex)
	pdf.SetDrawColor(navyR, navyG, navyB)
	pdf.SetLineWidth(ptToMM(0.5))
	pdf.Rect(cardPhotoX, cardPhotoY, cardPhotoW, cardPhotoH, "D")

	if registerOptionalImage(pdf, "card-photo", photoImg) {
		pdf.ImageOptions("card-photo", cardPhotoX, cardPhotoY, cardPhotoW, cardPhotoH, false, gofpdf.ImageOptions{}, 0, "")
		return
	}
	drawCardPhotoPlaceholder(pdf, cardPhotoX, cardPhotoY, cardPhotoW, cardPhotoH)
}

// drawCardPhotoPlaceholder draws a neutral silhouette (head + shoulders)
// clipped to the photo frame, so a NULL PhotoURL still leaves the frame
// occupying the identical position as the photo case (FR-21).
func drawCardPhotoPlaceholder(pdf *gofpdf.Fpdf, x, y, w, h float64) {
	r, g, b := hexRGB(cardPlaceholderHex)
	pdf.SetFillColor(r, g, b)
	cx := x + w/2

	pdf.ClipRect(x, y, w, h, false)
	pdf.Ellipse(cx, y+h*0.30, w*0.24, h*0.16, 0, "F")
	pdf.Ellipse(cx, y+h*1.15, w*0.55, h*0.5, 0, "F")
	pdf.ClipEnd()
}

const (
	cardDetailX     = 36.0
	cardDetailRight = 142.0
)

func drawCardDetailColumn(pdf *gofpdf.Fpdf, reg *model.RegistrationDetail, studentName string) {
	navyR, navyG, navyB := hexRGB(cardNavyHex)
	w := cardDetailRight - cardDetailX
	y := 24.0

	drawLabel := func(label string) {
		pdf.SetTextColor(navyR, navyG, navyB)
		pdf.SetFont(FontPublicSans, "", 6)
		pdf.SetXY(cardDetailX, y)
		pdf.CellFormat(w, 3, label, "", 0, "L", false, 0, "")
		y += 3.2
	}

	name := studentName
	if name == "" {
		name = "-"
	}
	drawLabel("NAMA")
	pdf.SetFont(FontSourceSerif4, "B", 10)
	pdf.SetXY(cardDetailX, y)
	pdf.CellFormat(w, 5, fitOneLine(pdf, name, w), "", 0, "L", false, 0, "")
	y += 6.5

	drawLabel("UJIAN")
	pdf.SetFont(FontPublicSans, "", 9)
	for _, line := range wrapLines(pdf, reg.Exam.Title, w, 2) {
		pdf.SetXY(cardDetailX, y)
		pdf.CellFormat(w, 4, line, "", 0, "L", false, 0, "")
		y += 4
	}
	y += 1.5

	drawLabel("JADWAL")
	pdf.SetFont(FontPublicSans, "", 8)
	pdf.SetXY(cardDetailX, y)
	pdf.CellFormat(w, 4, fitOneLine(pdf, cardScheduleText(reg), w), "", 0, "L", false, 0, "")
}

// cardScheduleText preserves the pre-existing schedule formatting: Asia/Jakarta,
// "02 Jan 2006 15:04 WIB" (FR-23).
func cardScheduleText(reg *model.RegistrationDetail) string {
	if reg.Exam.ScheduledAt == nil {
		return "-"
	}
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}
	return reg.Exam.ScheduledAt.In(loc).Format("02 Jan 2006 15:04 WIB")
}

const (
	cardTokenX = 8.0
	cardTokenY = 64.0
	cardTokenW = 132.0
	cardTokenH = 14.0
)

// drawCardTokenBand never truncates or wraps the token (FR-22, Invariant 5):
// only the font size shrinks to fit, the token string itself is never cut.
func drawCardTokenBand(pdf *gofpdf.Fpdf, token string) {
	bgR, bgG, bgB := hexRGB(cardGoldBgHex)
	navyR, navyG, navyB := hexRGB(cardNavyHex)

	pdf.SetFillColor(bgR, bgG, bgB)
	pdf.Rect(cardTokenX, cardTokenY, cardTokenW, cardTokenH, "F")

	pdf.SetTextColor(navyR, navyG, navyB)
	pdf.SetFont(FontPublicSans, "", 6)
	pdf.SetXY(cardTokenX, cardTokenY+1.5)
	pdf.CellFormat(cardTokenW, 3, "TOKEN", "", 0, "C", false, 0, "")

	size := shrinkToFit(pdf, FontSourceSerif4, "B", token, cardTokenW-12, 20, 8)
	pdf.SetFont(FontSourceSerif4, "B", size)
	pdf.SetXY(cardTokenX, cardTokenY+5.5)
	pdf.CellFormat(cardTokenW, 8, token, "", 0, "C", false, 0, "")
}

const (
	cardFooterX = 8.0
	cardFooterY = 82.0
	cardFooterW = 132.0
)

// drawCardFooterNote preserves the pre-existing check-in vs free-access copy,
// keyed on reg.Exam.RequiresCheckin / CheckInWindowMinutes. The bundled font
// set (Task 2) has no italic weight, so this renders in the regular style
// rather than the spec's "sans italic".
func drawCardFooterNote(pdf *gofpdf.Fpdf, reg *model.RegistrationDetail) {
	navyR, navyG, navyB := hexRGB(cardNavyHex)
	pdf.SetTextColor(navyR, navyG, navyB)
	pdf.SetFont(FontPublicSans, "", 6.5)
	pdf.SetXY(cardFooterX, cardFooterY)

	var note string
	if reg.Exam.RequiresCheckin {
		if reg.Exam.CheckInWindowMinutes != nil {
			note = fmt.Sprintf("Harap check-in dalam waktu %d menit sebelum ujian.", *reg.Exam.CheckInWindowMinutes)
		} else {
			note = "Harap check-in sebelum ujian dimulai."
		}
	} else {
		note = "Akses bebas pada waktu yang ditentukan."
	}
	pdf.MultiCell(cardFooterW, 3.2, note, "", "L", false)
}

func drawCardBorder(pdf *gofpdf.Fpdf) {
	navyR, navyG, navyB := hexRGB(cardNavyHex)
	pdf.SetDrawColor(navyR, navyG, navyB)
	pdf.SetLineWidth(ptToMM(0.6))
	pdf.RoundedRect(4, 4, cardPageW-8, cardPageH-8, 3, "1234", "D")
}

// registerOptionalImage validates and registers image bytes for placement,
// never leaving the pdf in an error state on bad input: a missing or corrupt
// asset must not fail card generation (FR-21). It returns false for empty,
// undecodable, or otherwise-rejected data.
func registerOptionalImage(pdf *gofpdf.Fpdf, name string, data []byte) bool {
	if len(data) == 0 {
		return false
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width == 0 || cfg.Height == 0 {
		return false
	}
	tp, ok := gofpdfImageType(format)
	if !ok {
		return false
	}
	pdf.RegisterImageOptionsReader(name, gofpdf.ImageOptions{ImageType: tp, ReadDpi: false}, bytes.NewReader(data))
	if !pdf.Ok() {
		pdf.ClearError()
		return false
	}
	return true
}

func gofpdfImageType(format string) (string, bool) {
	switch format {
	case "png":
		return "png", true
	case "jpeg":
		return "jpg", true
	case "gif":
		return "gif", true
	default:
		return "", false
	}
}

func ptToMM(pt float64) float64 {
	return pt * 25.4 / 72
}

func hexRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// shrinkToFit reduces font size in 0.5pt steps until text fits maxWidth,
// stopping at minSize. It never modifies text — callers that must guarantee
// a string is never truncated (e.g. the token, FR-22) use this instead of
// any truncating helper.
func shrinkToFit(pdf *gofpdf.Fpdf, family, style, text string, maxWidth, startSize, minSize float64) float64 {
	size := startSize
	for size > minSize {
		pdf.SetFont(family, style, size)
		if pdf.GetStringWidth(text) <= maxWidth {
			return size
		}
		size -= 0.5
	}
	pdf.SetFont(family, style, minSize)
	return minSize
}

// fitOneLine returns text unchanged if it fits maxWidth at the current font,
// otherwise truncates with an ellipsis so it never overflows.
func fitOneLine(pdf *gofpdf.Fpdf, text string, maxWidth float64) string {
	if pdf.GetStringWidth(text) <= maxWidth {
		return text
	}
	return truncateWithEllipsis(pdf, text, maxWidth)
}

func truncateWithEllipsis(pdf *gofpdf.Fpdf, text string, maxWidth float64) string {
	runes := []rune(text)
	for i := len(runes); i > 0; i-- {
		candidate := strings.TrimRight(string(runes[:i]), " ") + "…"
		if pdf.GetStringWidth(candidate) <= maxWidth {
			return candidate
		}
	}
	return "…"
}

// wrapLines word-wraps text at the current font to maxWidth, keeping at most
// maxLines; any remaining content is signalled by truncating the last kept
// line with an ellipsis rather than overflowing the border.
func wrapLines(pdf *gofpdf.Fpdf, text string, maxWidth float64, maxLines int) []string {
	raw := pdf.SplitLines([]byte(text), maxWidth)
	lines := make([]string, 0, maxLines)
	for i, l := range raw {
		if i >= maxLines {
			break
		}
		lines = append(lines, string(l))
	}
	if len(lines) == 0 {
		return []string{""}
	}
	if len(raw) > maxLines {
		lines[maxLines-1] = truncateWithEllipsis(pdf, lines[maxLines-1], maxWidth)
	}
	return lines
}
