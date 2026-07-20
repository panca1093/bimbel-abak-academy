package service

// Coordinate contract (FR-1, FR-2): every coordinate in this file is expressed in
// millimetres, origin at the page's top-left corner, Y increasing downward — no
// flip anywhere. x_mm,y_mm is the box's top-left corner and align applies inside
// the box. The renderer passes these values straight to gofpdf's SetXY in "mm"
// mode; the editor's only conversion is the uniform scale
// mm = px * (page_width_mm / preview_width_px). Never compute pageHeight - y.

import "fmt"

// FieldID identifies a certificate layout field. It aliases string rather than
// defining a distinct type so it drops in wherever LayoutField.ID and
// validLayoutFieldIDs already use plain strings, with no conversion at call sites.
type FieldID = string

// certificatePageWidthMm and certificatePageHeightMm are the A4 landscape page
// dimensions used by every certificate layout.
const (
	certificatePageWidthMm  = 297
	certificatePageHeightMm = 210
)

// validLayoutFieldIDs is the closed set from FR-3. Any other id is rejected.
var validLayoutFieldIDs = map[string]bool{
	"title":              true,
	"subtitle":           true,
	"student_name":       true,
	"exam_title":         true,
	"completion_text":    true,
	"date":               true,
	"certificate_number": true,
	"logo":               true,
}

// Page is the layout's page size in millimetres.
type Page struct {
	WidthMm  float64 `json:"width_mm"`
	HeightMm float64 `json:"height_mm"`
}

// Background describes the certificate's background image. Kind is "builtin"
// (Ref names one of classic/modern/elegant) or "custom" (Ref is an uploaded
// object key).
type Background struct {
	Kind string `json:"kind"`
	Ref  string `json:"ref"`
}

// LayoutField positions one stamped field. XMm/YMm is the box's top-left corner
// (FR-2); Align governs text placement inside the box. Font is validated against
// the closed FR-7a family set only at render time, not here: an unknown family
// must still render by falling back to a brand default (Invariant 8), so parse-time
// rejection would defeat that fallback. The logo field carries HMm instead of the
// text properties (Font/Weight/SizePt) — those are left zero-valued for it.
type LayoutField struct {
	ID      string  `json:"id"`
	XMm     float64 `json:"x_mm"`
	YMm     float64 `json:"y_mm"`
	WMm     float64 `json:"w_mm"`
	Align   string  `json:"align"`
	Font    string  `json:"font,omitempty"`
	Weight  string  `json:"weight,omitempty"`
	SizePt  float64 `json:"size_pt,omitempty"`
	Color   string  `json:"color,omitempty"`
	Visible bool    `json:"visible"`
	HMm     float64 `json:"h_mm,omitempty"`
}

// Layout is the JSON contract shared by the certificate renderer and the admin
// editor, persisted in exam.certificate_layout.
type Layout struct {
	Page       Page          `json:"page"`
	Background Background    `json:"background"`
	Fields     []LayoutField `json:"fields"`
}

// ValidateLayout rejects an unknown field id, a duplicate field id, and any field
// box that falls outside the page. It does not validate Font (see LayoutField).
func ValidateLayout(l Layout) error {
	seen := make(map[string]bool, len(l.Fields))
	for _, f := range l.Fields {
		if !validLayoutFieldIDs[f.ID] {
			return fmt.Errorf("%w: unknown field id: %s", ErrValidation, f.ID)
		}
		if seen[f.ID] {
			return fmt.Errorf("%w: duplicate field id: %s", ErrValidation, f.ID)
		}
		seen[f.ID] = true

		if f.XMm < 0 || f.YMm < 0 || f.XMm+f.WMm > l.Page.WidthMm || f.YMm > l.Page.HeightMm {
			return fmt.Errorf("%w: field %s box is outside the page", ErrValidation, f.ID)
		}
		if f.ID == "logo" && f.YMm+f.HMm > l.Page.HeightMm {
			return fmt.Errorf("%w: field %s box is outside the page", ErrValidation, f.ID)
		}
	}
	return nil
}

// defaultLayout returns the built-in default field placement for template
// ("classic", "modern", or "elegant"). All three share the A4 landscape page and
// the closed field id set; they differ in font, color, and layout rhythm.
func defaultLayout(template string) Layout {
	page := Page{WidthMm: certificatePageWidthMm, HeightMm: certificatePageHeightMm}

	switch template {
	case "modern":
		return Layout{
			Page:       page,
			Background: Background{Kind: "builtin", Ref: "modern"},
			Fields: []LayoutField{
				{ID: "title", XMm: 48.5, YMm: 42, WMm: 200, Align: "center", Font: "playfair_display", Weight: "bold", SizePt: 30, Color: "#0F172A", Visible: true},
				{ID: "subtitle", XMm: 48.5, YMm: 75, WMm: 200, Align: "center", Font: "public_sans", Weight: "regular", SizePt: 13, Color: "#334155", Visible: true},
				{ID: "student_name", XMm: 48.5, YMm: 98, WMm: 200, Align: "center", Font: "cormorant_garamond", Weight: "bold", SizePt: 32, Color: "#0F172A", Visible: true},
				{ID: "completion_text", XMm: 48.5, YMm: 118, WMm: 200, Align: "center", Font: "public_sans", Weight: "regular", SizePt: 12, Color: "#334155", Visible: true},
				{ID: "exam_title", XMm: 48.5, YMm: 133, WMm: 200, Align: "center", Font: "source_serif_4", Weight: "regular", SizePt: 15, Color: "#0F172A", Visible: true},
				{ID: "date", XMm: 48.5, YMm: 158, WMm: 200, Align: "center", Font: "public_sans", Weight: "regular", SizePt: 11, Color: "#334155", Visible: true},
				{ID: "certificate_number", XMm: 48.5, YMm: 196, WMm: 200, Align: "center", Font: "public_sans", Weight: "regular", SizePt: 9, Color: "#64748B", Visible: true},
				{ID: "logo", XMm: 138.5, YMm: 14, WMm: 20, Align: "center", Color: "", Visible: true, HMm: 20},
			},
		}
	case "elegant":
		return Layout{
			Page:       page,
			Background: Background{Kind: "builtin", Ref: "elegant"},
			Fields: []LayoutField{
				{ID: "title", XMm: 48.5, YMm: 44, WMm: 200, Align: "center", Font: "cinzel", Weight: "bold", SizePt: 28, Color: "#22315B", Visible: true},
				{ID: "subtitle", XMm: 48.5, YMm: 78, WMm: 200, Align: "center", Font: "cormorant_garamond", Weight: "regular", SizePt: 14, Color: "#4A5568", Visible: true},
				{ID: "student_name", XMm: 48.5, YMm: 102, WMm: 200, Align: "center", Font: "great_vibes", Weight: "regular", SizePt: 38, Color: "#22315B", Visible: true},
				{ID: "completion_text", XMm: 48.5, YMm: 122, WMm: 200, Align: "center", Font: "cormorant_garamond", Weight: "regular", SizePt: 12, Color: "#4A5568", Visible: true},
				{ID: "exam_title", XMm: 48.5, YMm: 137, WMm: 200, Align: "center", Font: "source_serif_4", Weight: "regular", SizePt: 15, Color: "#22315B", Visible: true},
				{ID: "date", XMm: 48.5, YMm: 160, WMm: 200, Align: "center", Font: "cormorant_garamond", Weight: "regular", SizePt: 12, Color: "#4A5568", Visible: true},
				{ID: "certificate_number", XMm: 48.5, YMm: 197, WMm: 200, Align: "center", Font: "public_sans", Weight: "regular", SizePt: 9, Color: "#8A93A6", Visible: true},
				{ID: "logo", XMm: 138.5, YMm: 15, WMm: 20, Align: "center", Color: "", Visible: true, HMm: 20},
			},
		}
	default: // "classic"
		return Layout{
			Page:       page,
			Background: Background{Kind: "builtin", Ref: "classic"},
			Fields: []LayoutField{
				{ID: "title", XMm: 48.5, YMm: 42, WMm: 200, Align: "center", Font: "source_serif_4", Weight: "bold", SizePt: 28, Color: "#1F2A44", Visible: true},
				{ID: "subtitle", XMm: 48.5, YMm: 76, WMm: 200, Align: "center", Font: "public_sans", Weight: "regular", SizePt: 13, Color: "#4A5568", Visible: true},
				{ID: "student_name", XMm: 48.5, YMm: 100, WMm: 200, Align: "center", Font: "source_serif_4", Weight: "bold", SizePt: 26, Color: "#1F2A44", Visible: true},
				{ID: "completion_text", XMm: 48.5, YMm: 120, WMm: 200, Align: "center", Font: "public_sans", Weight: "regular", SizePt: 12, Color: "#4A5568", Visible: true},
				{ID: "exam_title", XMm: 48.5, YMm: 135, WMm: 200, Align: "center", Font: "source_serif_4", Weight: "regular", SizePt: 15, Color: "#1F2A44", Visible: true},
				{ID: "date", XMm: 48.5, YMm: 158, WMm: 200, Align: "center", Font: "public_sans", Weight: "regular", SizePt: 11, Color: "#4A5568", Visible: true},
				{ID: "certificate_number", XMm: 48.5, YMm: 195, WMm: 200, Align: "center", Font: "public_sans", Weight: "regular", SizePt: 9, Color: "#8A93A6", Visible: true},
				{ID: "logo", XMm: 138.5, YMm: 15, WMm: 20, Align: "center", Color: "", Visible: true, HMm: 20},
			},
		}
	}
}
