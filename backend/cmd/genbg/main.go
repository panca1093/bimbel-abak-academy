// Command genbg draws the three built-in certificate backgrounds (classic, modern,
// elegant) as one-page A4-landscape PDFs, per spec OQ3. `make gen-cert-backgrounds`
// runs this then rasterizes each PDF to a committed PNG with pdftoppm. Artwork only —
// no text is drawn here; certificate text is stamped at render time from the layout
// schema.
//
// Design direction: the three templates are three registers of one Abak identity,
// rooted in Minangkabau songket — West Sumatra's gold-thread ceremonial weaving —
// since "Abak" is the Minangkabau word for father and the brand is family + heritage.
// The recurring signature is a gold songket diamond-weave band (drawSongketH /
// drawSongketV); the mark is the parent-and-child figure (drawFamilyMark).
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
)

const (
	pageWidthMM = 297.0
	// pageHeightMM is nudged 0.07mm under true A4 (210.0mm). At exact 210mm,
	// pdftoppm -r 150 rasterizes to 1241px tall (poppler ceils page-height-in-points
	// * 150/72, and 210mm's point value crosses the 1240px boundary by a hair);
	// this keeps the page comfortably inside the interval that ceils to exactly
	// 1240px per OQ3, with margin against floating-point noise. Imperceptible on
	// a background image.
	pageHeightMM = 209.93
)

// fixedTimestamp pins gofpdf's CreationDate/ModDate so `make gen-cert-backgrounds`
// produces byte-identical PDFs (and therefore PNGs) across runs; gofpdf otherwise
// embeds time.Now().
var fixedTimestamp = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

var templateNames = []string{"classic", "modern", "elegant"}

var templateDrawers = map[string]func(*gofpdf.Fpdf){
	"classic": drawClassic,
	"modern":  drawModern,
	"elegant": drawElegant,
}

func main() {
	outDir := "internal/service/assets"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "genbg:", err)
		os.Exit(1)
	}

	for _, name := range templateNames {
		pdf := newPage()
		templateDrawers[name](pdf)
		out := filepath.Join(outDir, fmt.Sprintf("cert_bg_%s.pdf", name))
		if err := pdf.OutputFileAndClose(out); err != nil {
			fmt.Fprintln(os.Stderr, "genbg:", err)
			os.Exit(1)
		}
	}
}

func newPage() *gofpdf.Fpdf {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "L",
		UnitStr:        "mm",
		Size:           gofpdf.SizeType{Wd: pageHeightMM, Ht: pageWidthMM},
	})
	pdf.SetCreationDate(fixedTimestamp)
	pdf.SetModificationDate(fixedTimestamp)
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	return pdf
}

// rgb is a brand palette color (design-app-abak/logo/logo-tokens.css).
type rgb struct{ r, g, b int }

var (
	navy700 = rgb{0x22, 0x31, 0x5B}
	navy500 = rgb{0x2E, 0x41, 0x74}
	gold600 = rgb{0xC6, 0x88, 0x1F}
	gold500 = rgb{0xD9, 0x9A, 0x2B}
	gold300 = rgb{0xF0, 0xCB, 0x78}
	gold100 = rgb{0xFB, 0xEF, 0xD4}
	teal600 = rgb{0x15, 0x7A, 0x6E}
	teal500 = rgb{0x1E, 0x97, 0x8A}
	paper   = rgb{0xFB, 0xFA, 0xF6}
	paper2  = rgb{0xF4, 0xF1, 0xE9}
	white   = rgb{0xFF, 0xFF, 0xFF}
)

func setFill(pdf *gofpdf.Fpdf, c rgb) { pdf.SetFillColor(c.r, c.g, c.b) }
func setDraw(pdf *gofpdf.Fpdf, c rgb) { pdf.SetDrawColor(c.r, c.g, c.b) }

// --- Parent-and-child mark (design-app-abak/logo/mark-family-*.svg) ---
//
// Transcribed from the SVG's 120x120 viewBox. Body shapes are two quadratic
// beziers each (SVG `Q` commands map 1:1 to gofpdf CurveTo); heads are circles,
// the accent kite is a polygon, the stem a rect. Colors are passed in so the mark
// reads on both dark grounds (light parent) and light grounds (navy parent).
func drawFamilyMark(pdf *gofpdf.Fpdf, cx, cy, size float64, parentC, childC, accentC rgb) {
	s := size / 120.0
	ox, oy := cx-size/2, cy-size/2
	X := func(v float64) float64 { return ox + v*s }
	Y := func(v float64) float64 { return oy + v*s }

	// parent body — M22,104 Q22,64 44,64 Q66,64 66,104 Z
	setFill(pdf, parentC)
	pdf.MoveTo(X(22), Y(104))
	pdf.CurveTo(X(22), Y(64), X(44), Y(64))
	pdf.CurveTo(X(66), Y(64), X(66), Y(104))
	pdf.ClosePath()
	pdf.DrawPath("F")
	// parent head
	pdf.Circle(X(44), Y(34), 15*s, "F")

	// child body — M62,104 Q62,78 80,78 Q98,78 98,104 Z
	setFill(pdf, childC)
	pdf.MoveTo(X(62), Y(104))
	pdf.CurveTo(X(62), Y(78), X(80), Y(78))
	pdf.CurveTo(X(98), Y(78), X(98), Y(104))
	pdf.ClosePath()
	pdf.DrawPath("F")

	// accent kite — M80,44 L96,51 L80,58 L64,51 Z
	setFill(pdf, accentC)
	pdf.Polygon([]gofpdf.PointType{
		{X: X(80), Y: Y(44)}, {X: X(96), Y: Y(51)}, {X: X(80), Y: Y(58)}, {X: X(64), Y: Y(51)},
	}, "F")

	// child head (over the kite base)
	setFill(pdf, childC)
	pdf.Circle(X(80), Y(62), 11*s, "F")

	// gold stem
	setFill(pdf, accentC)
	pdf.Rect(X(79), Y(44), 2.5*s, 9*s, "F")
}

// drawSongketH draws a horizontal songket ribbon — a chain of gold lozenge "rings"
// (saik galamai diamond motif) centred on yMid across width w, framed by two thin
// accent rules. groundC is the band colour the inner diamonds are punched in.
func drawSongketH(pdf *gofpdf.Fpdf, cx, yMid, w float64, threadC, groundC, accentC rgb) {
	const pitch, amp = 9.0, 3.1
	n := int(w / pitch)
	if n < 1 {
		n = 1
	}
	startX := cx - float64(n)*pitch/2
	diamond := func(x, hh, hw float64, c rgb) {
		setFill(pdf, c)
		pdf.Polygon([]gofpdf.PointType{
			{X: x, Y: yMid - hh}, {X: x + hw, Y: yMid}, {X: x, Y: yMid + hh}, {X: x - hw, Y: yMid},
		}, "F")
	}
	for i := 0; i < n; i++ {
		x := startX + float64(i)*pitch + pitch/2
		diamond(x, amp, pitch/2, threadC)      // gold lozenge
		diamond(x, amp*0.5, pitch/4, groundC)  // punched centre → ring
	}
	// small solid gold pips at the joins
	setFill(pdf, threadC)
	for i := 0; i <= n; i++ {
		x := startX + float64(i)*pitch
		pdf.Polygon([]gofpdf.PointType{
			{X: x, Y: yMid - 1.0}, {X: x + 1.0, Y: yMid}, {X: x, Y: yMid + 1.0}, {X: x - 1.0, Y: yMid},
		}, "F")
	}
	setDraw(pdf, accentC)
	pdf.SetLineWidth(0.35)
	pdf.Line(startX, yMid-amp-1.6, startX+float64(n)*pitch, yMid-amp-1.6)
	pdf.Line(startX, yMid+amp+1.6, startX+float64(n)*pitch, yMid+amp+1.6)
}

// drawSongketV is the vertical counterpart — a lozenge chain running down xMid over
// height h, framed by two vertical accent rules.
func drawSongketV(pdf *gofpdf.Fpdf, xMid, cy, h float64, threadC, groundC, accentC rgb) {
	const pitch, amp = 9.0, 3.1
	n := int(h / pitch)
	if n < 1 {
		n = 1
	}
	startY := cy - float64(n)*pitch/2
	diamond := func(y, hh, hw float64, c rgb) {
		setFill(pdf, c)
		pdf.Polygon([]gofpdf.PointType{
			{X: xMid, Y: y - hh}, {X: xMid + hw, Y: y}, {X: xMid, Y: y + hh}, {X: xMid - hw, Y: y},
		}, "F")
	}
	for i := 0; i < n; i++ {
		y := startY + float64(i)*pitch + pitch/2
		diamond(y, pitch/2, amp, threadC)
		diamond(y, pitch/4, amp*0.5, groundC)
	}
	setFill(pdf, threadC)
	for i := 0; i <= n; i++ {
		y := startY + float64(i)*pitch
		pdf.Polygon([]gofpdf.PointType{
			{X: xMid, Y: y - 1.0}, {X: xMid + 1.0, Y: y}, {X: xMid, Y: y + 1.0}, {X: xMid - 1.0, Y: y},
		}, "F")
	}
	setDraw(pdf, accentC)
	pdf.SetLineWidth(0.35)
	pdf.Line(xMid-amp-1.6, startY, xMid-amp-1.6, startY+float64(n)*pitch)
	pdf.Line(xMid+amp+1.6, startY, xMid+amp+1.6, startY+float64(n)*pitch)
}

// cornerDiamond draws a small solid gold diamond centred at (x,y) — a woven-corner pin.
func cornerDiamond(pdf *gofpdf.Fpdf, x, y, r float64, c rgb) {
	setFill(pdf, c)
	pdf.Polygon([]gofpdf.PointType{
		{X: x, Y: y - r}, {X: x + r, Y: y}, {X: x, Y: y + r}, {X: x - r, Y: y},
	}, "F")
}

// drawClassic — flagship: navy songket header + footer bands on a warm cream field,
// the parent-and-child mark woven into the header. Ceremonial, gold-on-indigo.
func drawClassic(pdf *gofpdf.Fpdf) {
	setFill(pdf, paper2)
	pdf.Rect(0, 0, pageWidthMM, pageHeightMM, "F")

	// top navy band
	const topH = 47.0
	setFill(pdf, navy700)
	pdf.Rect(0, 0, pageWidthMM, topH, "F")
	drawFamilyMark(pdf, pageWidthMM/2, 15.5, 21, paper, teal500, gold500)
	drawSongketH(pdf, pageWidthMM/2, 35, pageWidthMM-70, gold300, navy700, gold500)
	// double gold rule under the band
	setFill(pdf, gold500)
	pdf.Rect(0, topH, pageWidthMM, 1.3, "F")
	setFill(pdf, gold300)
	pdf.Rect(0, topH+1.9, pageWidthMM, 0.45, "F")

	// bottom navy band
	const botH = 30.0
	botY := pageHeightMM - botH
	setFill(pdf, gold300)
	pdf.Rect(0, botY-1.9, pageWidthMM, 0.45, "F")
	setFill(pdf, gold500)
	pdf.Rect(0, botY-0.6, pageWidthMM, 1.3, "F")
	setFill(pdf, navy700)
	pdf.Rect(0, botY, pageWidthMM, botH, "F")
	// small woven pips flanking the certificate-number line (kept clear of the text)
	cornerDiamond(pdf, pageWidthMM/2-46, botY+botH/2, 1.6, gold500)
	cornerDiamond(pdf, pageWidthMM/2+46, botY+botH/2, 1.6, gold500)

	// thin gold keyline framing the cream field
	setDraw(pdf, gold500)
	pdf.SetLineWidth(0.3)
	pdf.Rect(15, topH+9, pageWidthMM-30, botY-topH-18, "D")
	cornerDiamond(pdf, 15, topH+9, 1.6, gold500)
	cornerDiamond(pdf, pageWidthMM-15, topH+9, 1.6, gold500)
	cornerDiamond(pdf, 15, botY-9, 1.6, gold500)
	cornerDiamond(pdf, pageWidthMM-15, botY-9, 1.6, gold500)
}

// drawModern — clean and asymmetric: a navy songket band down the left edge anchors
// a generous cream text field. Fixes the previous sparse/unbalanced look.
func drawModern(pdf *gofpdf.Fpdf) {
	setFill(pdf, paper)
	pdf.Rect(0, 0, pageWidthMM, pageHeightMM, "F")

	const bandW = 46.0
	setFill(pdf, navy700)
	pdf.Rect(0, 0, bandW, pageHeightMM, "F")
	drawFamilyMark(pdf, bandW/2, 26, 24, paper, teal500, gold500)
	drawSongketV(pdf, bandW/2, pageHeightMM/2+18, pageHeightMM-90, gold300, navy700, gold500)

	// teal keyline separating band from field
	setDraw(pdf, teal500)
	pdf.SetLineWidth(1.1)
	pdf.Line(bandW+3, 0, bandW+3, pageHeightMM)
	setDraw(pdf, gold500)
	pdf.SetLineWidth(0.4)
	pdf.Line(bandW+5, 0, bandW+5, pageHeightMM)

	// a short gold rule under the title zone, in the field (asymmetric accent)
	fieldCx := (bandW + pageWidthMM) / 2
	setDraw(pdf, gold500)
	pdf.SetLineWidth(0.6)
	pdf.Line(fieldCx-24, 64, fieldCx+24, 64)
	cornerDiamond(pdf, fieldCx, 64, 1.4, gold500)
}

// drawElegant — heirloom: warm paper with a fine gold songket border ribbon top and
// bottom, slim gold side rules, and the parent-and-child mark in a gold medallion.
func drawElegant(pdf *gofpdf.Fpdf) {
	setFill(pdf, paper)
	pdf.Rect(0, 0, pageWidthMM, pageHeightMM, "F")

	const inset = 14.0
	// songket ribbons top and bottom
	drawSongketH(pdf, pageWidthMM/2, inset, pageWidthMM-2*inset-14, gold500, paper, gold600)
	drawSongketH(pdf, pageWidthMM/2, pageHeightMM-inset, pageWidthMM-2*inset-14, gold500, paper, gold600)
	// slim gold rules down the sides, tying the ribbons together
	setDraw(pdf, gold500)
	pdf.SetLineWidth(0.4)
	pdf.Line(inset, inset, inset, pageHeightMM-inset)
	pdf.Line(pageWidthMM-inset, inset, pageWidthMM-inset, pageHeightMM-inset)
	setDraw(pdf, gold300)
	pdf.SetLineWidth(0.3)
	pdf.Line(inset+2, inset, inset+2, pageHeightMM-inset)
	pdf.Line(pageWidthMM-inset-2, inset, pageWidthMM-inset-2, pageHeightMM-inset)
	// woven-corner pins
	cornerDiamond(pdf, inset, inset, 2.1, gold600)
	cornerDiamond(pdf, pageWidthMM-inset, inset, 2.1, gold600)
	cornerDiamond(pdf, inset, pageHeightMM-inset, 2.1, gold600)
	cornerDiamond(pdf, pageWidthMM-inset, pageHeightMM-inset, 2.1, gold600)

	// medallion with the true-colour family mark
	const medY = 34.0
	setFill(pdf, gold100)
	pdf.Circle(pageWidthMM/2, medY, 15, "F")
	setDraw(pdf, gold600)
	pdf.SetLineWidth(0.5)
	pdf.Circle(pageWidthMM/2, medY, 15, "D")
	setDraw(pdf, gold300)
	pdf.SetLineWidth(0.3)
	pdf.Circle(pageWidthMM/2, medY, 12.4, "D")
	drawFamilyMark(pdf, pageWidthMM/2, medY, 19, navy700, teal500, gold600)
	// hairline rules flanking the medallion
	setDraw(pdf, gold500)
	pdf.SetLineWidth(0.3)
	pdf.Line(inset+18, medY, pageWidthMM/2-19, medY)
	pdf.Line(pageWidthMM/2+19, medY, pageWidthMM-inset-18, medY)
}
