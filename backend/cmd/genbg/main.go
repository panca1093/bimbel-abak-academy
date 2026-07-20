// Command genbg draws the three built-in certificate backgrounds (classic, modern,
// elegant) as one-page A4-landscape PDFs, per spec OQ3. `make gen-cert-backgrounds`
// runs this then rasterizes each PDF to a committed PNG with pdftoppm. Artwork only —
// no text is drawn here; certificate text is stamped at render time from the layout
// schema.
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
	gold600 = rgb{0xC6, 0x88, 0x1F}
	gold500 = rgb{0xD9, 0x9A, 0x2B}
	gold300 = rgb{0xF0, 0xCB, 0x78}
	gold100 = rgb{0xFB, 0xEF, 0xD4}
	teal600 = rgb{0x15, 0x7A, 0x6E}
	teal500 = rgb{0x1E, 0x97, 0x8A}
	teal100 = rgb{0xDC, 0xF0, 0xEC}
	paper   = rgb{0xFB, 0xFA, 0xF6}
	paper2  = rgb{0xF3, 0xF1, 0xE9}
	white   = rgb{0xFF, 0xFF, 0xFF}
)

func setFill(pdf *gofpdf.Fpdf, c rgb) { pdf.SetFillColor(c.r, c.g, c.b) }
func setDraw(pdf *gofpdf.Fpdf, c rgb) { pdf.SetDrawColor(c.r, c.g, c.b) }

// --- Gonjong mark (design-app-abak/logo/mark-gonjong-*.svg, OQ1) ---
//
// Point data is transcribed from the SVG's 120x120 viewBox path commands. The
// roofline path's one `L34 58` segment is expressed as a degenerate cubic bezier
// (control points equal to its endpoints) so the whole outline fits gofpdf's
// Beziergon, which only accepts curve segments.

type pt struct{ x, y float64 }

const gonjongViewBox = 120.0

var gonjongBody = []pt{{36, 96}, {40, 58}, {80, 58}, {84, 96}}

var gonjongRoof = []pt{
	{8, 14},
	{24, 30}, {44, 44}, {60, 46},
	{76, 44}, {96, 30}, {112, 14},
	{104, 30}, {95, 50}, {86, 58},
	{86, 58}, {34, 58}, {34, 58}, // degenerate bezier == straight line
	{25, 50}, {16, 30}, {8, 14},
}

const (
	gonjongDoorX, gonjongDoorY = 55.0, 78.0
	gonjongDoorW, gonjongDoorH = 10.0, 18.0
	gonjongDoorR               = 2.0
)

// drawGonjongMark draws the mark centered at (cx, cy) mm with bounding size sizeMM,
// using bodyColor for the house silhouette and accentColor for the roof and door.
func drawGonjongMark(pdf *gofpdf.Fpdf, cx, cy, sizeMM float64, bodyColor, accentColor rgb) {
	scale := sizeMM / gonjongViewBox
	ox, oy := cx-sizeMM/2, cy-sizeMM/2
	tr := func(p pt) gofpdf.PointType {
		return gofpdf.PointType{X: ox + p.x*scale, Y: oy + p.y*scale}
	}

	setFill(pdf, bodyColor)
	body := make([]gofpdf.PointType, len(gonjongBody))
	for i, p := range gonjongBody {
		body[i] = tr(p)
	}
	pdf.Polygon(body, "F")

	setFill(pdf, accentColor)
	roof := make([]gofpdf.PointType, len(gonjongRoof))
	for i, p := range gonjongRoof {
		roof[i] = tr(p)
	}
	pdf.Beziergon(roof, "F")

	door := tr(pt{gonjongDoorX, gonjongDoorY})
	pdf.RoundedRect(door.X, door.Y, gonjongDoorW*scale, gonjongDoorH*scale, gonjongDoorR*scale, "1234", "F")
}

// drawCornerBrackets draws L-shaped brackets at the four corners of rect (x,y,w,h),
// arm length arm mm.
func drawCornerBrackets(pdf *gofpdf.Fpdf, x, y, w, h float64, c rgb, arm float64) {
	setDraw(pdf, c)
	pdf.SetLineWidth(0.5)
	x2, y2 := x+w, y+h
	pdf.Line(x, y, x+arm, y)
	pdf.Line(x, y, x, y+arm)
	pdf.Line(x2, y, x2-arm, y)
	pdf.Line(x2, y, x2, y+arm)
	pdf.Line(x, y2, x+arm, y2)
	pdf.Line(x, y2, x, y2-arm)
	pdf.Line(x2, y2, x2-arm, y2)
	pdf.Line(x2, y2, x2, y2-arm)
}

// drawCornerFlourish draws a single quarter-curve tucked into the corner (x,y),
// bulging in direction (sx, sy) (each ±1).
func drawCornerFlourish(pdf *gofpdf.Fpdf, x, y, sx, sy float64, c rgb) {
	setDraw(pdf, c)
	pdf.SetLineWidth(0.5)
	pdf.Curve(x+sx*18, y, x+sx*6, y+sy*6, x, y+sy*18, "D")
}

// drawClassic: navy header/footer bands, double-rule frame, corner brackets — blue,
// formal (matches classicLayout's color intent in certificate.go).
func drawClassic(pdf *gofpdf.Fpdf) {
	setFill(pdf, paper)
	pdf.Rect(0, 0, pageWidthMM, pageHeightMM, "F")

	setDraw(pdf, navy700)
	pdf.SetLineWidth(0.6)
	pdf.Rect(10, 40, pageWidthMM-20, pageHeightMM-80, "D")
	setDraw(pdf, gold500)
	pdf.SetLineWidth(0.3)
	pdf.Rect(13, 43, pageWidthMM-26, pageHeightMM-86, "D")

	drawCornerBrackets(pdf, 16, 46, pageWidthMM-32, pageHeightMM-92, navy700, 6)

	const bandH = 32.0
	const ruleH = 1.2
	setFill(pdf, navy700)
	pdf.Rect(0, 0, pageWidthMM, bandH, "F")
	setFill(pdf, gold500)
	pdf.Rect(0, bandH, pageWidthMM, ruleH, "F")

	setFill(pdf, gold500)
	pdf.Rect(0, pageHeightMM-bandH-ruleH, pageWidthMM, ruleH, "F")
	setFill(pdf, navy700)
	pdf.Rect(0, pageHeightMM-bandH, pageWidthMM, bandH, "F")

	drawGonjongMark(pdf, pageWidthMM/2, bandH/2, 18, white, gold300)
}

// drawModern: white ground, teal accent stripes and single-rule frame, teal badge —
// clean (matches modernLayout's color intent in certificate.go).
func drawModern(pdf *gofpdf.Fpdf) {
	setFill(pdf, white)
	pdf.Rect(0, 0, pageWidthMM, pageHeightMM, "F")

	setFill(pdf, teal500)
	pdf.Rect(0, 0, pageWidthMM, 4, "F")
	pdf.Rect(0, pageHeightMM-4, pageWidthMM, 4, "F")

	setDraw(pdf, teal500)
	pdf.SetLineWidth(0.4)
	pdf.Rect(10, 10, pageWidthMM-20, pageHeightMM-20, "D")

	setFill(pdf, teal100)
	pdf.RoundedRect(14, 12, 26, 26, 4, "1234", "F")
	drawGonjongMark(pdf, 14+13, 12+13, 18, navy700, teal600)
}

// drawElegant: cream ground, ornate double-rule border with corner flourishes, gold
// medallion — matches elegantLayout's color intent in certificate.go.
func drawElegant(pdf *gofpdf.Fpdf) {
	setFill(pdf, paper2)
	pdf.Rect(0, 0, pageWidthMM, pageHeightMM, "F")

	setDraw(pdf, gold600)
	pdf.SetLineWidth(0.7)
	pdf.Rect(8, 8, pageWidthMM-16, pageHeightMM-16, "D")
	setDraw(pdf, gold500)
	pdf.SetLineWidth(0.3)
	pdf.Rect(11, 11, pageWidthMM-22, pageHeightMM-22, "D")

	drawCornerFlourish(pdf, 8, 8, 1, 1, gold500)
	drawCornerFlourish(pdf, pageWidthMM-8, 8, -1, 1, gold500)
	drawCornerFlourish(pdf, 8, pageHeightMM-8, 1, -1, gold500)
	drawCornerFlourish(pdf, pageWidthMM-8, pageHeightMM-8, -1, -1, gold500)

	const medallionY = 30.0
	setFill(pdf, gold100)
	pdf.Circle(pageWidthMM/2, medallionY, 15, "F")
	setDraw(pdf, gold500)
	pdf.SetLineWidth(0.4)
	pdf.Circle(pageWidthMM/2, medallionY, 15, "D")
	drawGonjongMark(pdf, pageWidthMM/2, medallionY, 20, navy700, gold600)

	setDraw(pdf, gold500)
	pdf.SetLineWidth(0.3)
	pdf.Line(30, medallionY, pageWidthMM/2-16, medallionY)
	pdf.Line(pageWidthMM/2+16, medallionY, pageWidthMM-30, medallionY)
}
