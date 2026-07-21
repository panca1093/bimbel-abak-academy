package service

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// registerOptionalImage validates and registers image bytes for placement,
// never leaving the pdf in an error state on bad input: a missing or corrupt
// asset must not fail card generation (FR-21). It returns ok=false for empty,
// undecodable, or otherwise-rejected data. width/height are the source
// image's own pixel dimensions (0,0 when ok is false), letting a caller
// aspect-fill the image into a box instead of stretching it.
func registerOptionalImage(pdf *gofpdf.Fpdf, name string, data []byte) (ok bool, width, height int) {
	if len(data) == 0 {
		return false, 0, 0
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width == 0 || cfg.Height == 0 {
		return false, 0, 0
	}
	tp, supported := gofpdfImageType(format)
	if !supported {
		return false, 0, 0
	}
	pdf.RegisterImageOptionsReader(name, gofpdf.ImageOptions{ImageType: tp, ReadDpi: false}, bytes.NewReader(data))
	if !pdf.Ok() {
		pdf.ClearError()
		return false, 0, 0
	}
	return true, cfg.Width, cfg.Height
}

// drawAspectFillImage draws a registered image scaled to cover a w x h mm box
// and centre-cropped to it via ClipRect, rather than stretched to fit — a
// non-matching source aspect ratio (e.g. a portrait phone photo on the exam
// card's 22x28mm frame, Warning 8) never squashes the subject.
func drawAspectFillImage(pdf *gofpdf.Fpdf, name string, x, y, w, h float64, srcW, srcH int) {
	boxAspect := w / h
	srcAspect := float64(srcW) / float64(srcH)

	drawW, drawH := w, h
	if srcAspect > boxAspect {
		drawH = h
		drawW = h * srcAspect
	} else {
		drawW = w
		drawH = w / srcAspect
	}
	drawX := x - (drawW-w)/2
	drawY := y - (drawH-h)/2

	pdf.ClipRect(x, y, w, h, false)
	pdf.ImageOptions(name, drawX, drawY, drawW, drawH, false, gofpdf.ImageOptions{}, 0, "")
	pdf.ClipEnd()
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
