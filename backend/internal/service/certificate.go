package service

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/minio/minio-go/v7"
)

// validCertificateTemplates is the closed set of accepted certificate layouts.
var validCertificateTemplates = map[string]bool{
	"classic": true,
	"modern":  true,
	"elegant": true,
	"custom":  true,
}

// shrinkToFitSafetyMargin leaves headroom below the exact fitted width so
// glyph rendering rounding doesn't tip a shrunk field back over the box edge.
// minShrinkToFitSizePt is the floor below which shrinking stops.
const (
	shrinkToFitSafetyMargin = 0.97
	minShrinkToFitSizePt    = 6.0
)

func validateCertificateTemplate(tmpl string) error {
	if !validCertificateTemplates[tmpl] {
		return fmt.Errorf("%w: unknown certificate template: %s", ErrValidation, tmpl)
	}
	return nil
}

//go:embed assets/cert_bg_classic.png
var certBgClassicPNG []byte

//go:embed assets/cert_bg_modern.png
var certBgModernPNG []byte

//go:embed assets/cert_bg_elegant.png
var certBgElegantPNG []byte

// builtinCertificateBackground returns the embedded background PNG for a
// built-in background ref. An unrecognised ref falls back to classic, which
// covers the "custom template but no background key" case (FR-19) since the
// caller passes the default layout's own ref in that situation.
func builtinCertificateBackground(ref string) []byte {
	switch ref {
	case "modern":
		return certBgModernPNG
	case "elegant":
		return certBgElegantPNG
	default:
		return certBgClassicPNG
	}
}

// renderCertificate is the single rendering entry point for real generation,
// preview, and the editor (FR-4): it draws bg full-bleed on an A4 landscape
// page, then stamps each visible field from layout with its value from vals.
// It performs no I/O — bg is supplied by the caller, embedded asset for
// built-ins or a downloaded object for custom (FR-8). Coordinates are passed
// straight to SetXY with no Y-axis arithmetic (FR-1).
func renderCertificate(layout Layout, bg []byte, vals map[FieldID]string) ([]byte, error) {
	// gofpdf.NewCustom treats Size as the *portrait* reference and swaps it
	// for OrientationStr "L" (see gofpdf fpdfNew). layout.Page is already
	// expressed in landscape terms, so orientation "P" applies it unswapped.
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size:           gofpdf.SizeType{Wd: layout.Page.WidthMm, Ht: layout.Page.HeightMm},
	})
	if err := RegisterFonts(pdf); err != nil {
		return nil, fmt.Errorf("register fonts: %w", err)
	}
	// Certificates are exactly one page by construction (FR-6); gofpdf's
	// default auto page-break would otherwise silently start a second page
	// when a field near the bottom margin is stamped.
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	imgOpt := gofpdf.ImageOptions{ImageType: "PNG"}
	pdf.RegisterImageOptionsReader("certificate-background", imgOpt, bytes.NewReader(bg))
	pdf.ImageOptions("certificate-background", 0, 0, layout.Page.WidthMm, layout.Page.HeightMm, false, imgOpt, 0, "")

	for _, field := range layout.Fields {
		if !field.Visible || field.ID == "logo" {
			continue
		}
		text := vals[field.ID]
		if text == "" {
			continue
		}

		style := ""
		if field.Weight == "bold" {
			style = "B"
		}
		family := ResolveFontFamily(field.Font)
		pdf.SetFont(family, style, field.SizePt)

		// Shrink-to-fit: CellFormat does not wrap or clip, so a long value
		// (e.g. a ~60-char name) at the field's nominal size can run past
		// the page edge (FR-6). Scale the font down to the field's box
		// width before drawing rather than let it overflow.
		size := field.SizePt
		if textWidth := pdf.GetStringWidth(text); textWidth > field.WMm && textWidth > 0 {
			size = field.SizePt * (field.WMm / textWidth) * shrinkToFitSafetyMargin
			if size < minShrinkToFitSizePt {
				size = minShrinkToFitSizePt
			}
			pdf.SetFont(family, style, size)
		}

		r, g, b := hexToRGB(field.Color)
		pdf.SetTextColor(r, g, b)

		pdf.SetXY(field.XMm, field.YMm)
		lineHeightMm := size * 0.3528 * 1.15
		pdf.CellFormat(field.WMm, lineHeightMm, text, "", 0, alignToGofpdf(field.Align), false, 0, "")
	}

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("render certificate: %w", err)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("output certificate pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// alignToGofpdf maps the layout schema's align values to gofpdf's CellFormat
// alignStr codes; an unrecognised value centers, matching the field default.
func alignToGofpdf(align string) string {
	switch align {
	case "left":
		return "L"
	case "right":
		return "R"
	default:
		return "C"
	}
}

// hexToRGB parses a "#RRGGBB" color into 0-255 components; a malformed value
// falls back to black rather than erroring, since layout field colors are not
// validated at parse time (see certificate_layout.go).
func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b int
	if len(hex) != 6 {
		return 0, 0, 0
	}
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return 0, 0, 0
	}
	return r, g, b
}

// certificateFieldValues assembles the fixed certificate copy plus the
// per-render student name, date, and certificate number shared by real
// generation, preview, and the standalone generateCertificatePDF helper.
func certificateFieldValues(examTitle, studentName, dateStr, certNumber string) map[FieldID]string {
	return map[FieldID]string{
		"title":              "CERTIFICATE OF COMPLETION",
		"subtitle":           "This certificate is proudly awarded to",
		"student_name":       studentName,
		"completion_text":    "for successfully completing",
		"exam_title":         examTitle,
		"date":               dateStr,
		"certificate_number": certNumber,
	}
}

// generateCertificatePDF renders a single-page A4 landscape certificate PDF
// through renderCertificate, using the given template's default layout and
// built-in background, with student name, exam title, and submission date
// (Asia/Jakarta, no score) stamped in.
func generateCertificatePDF(template, studentName, examTitle string, submittedAt time.Time) ([]byte, error) {
	if err := validateCertificateTemplate(template); err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	dateStr := submittedAt.In(loc).Format("2 January 2006")

	layout := defaultLayout(template)
	bg := builtinCertificateBackground(layout.Background.Ref)
	vals := certificateFieldValues(examTitle, studentName, dateStr, "ABK/2026/000000")

	return renderCertificate(layout, bg, vals)
}

// resolveCertificateLayout returns exam.CertificateLayout when the admin has
// saved one, else the built-in default layout for the exam's template (FR-29)
// — an exam never opens to an empty canvas. A "custom" template with no saved
// layout seeds from classic, mirroring the background fallback (FR-19).
func resolveCertificateLayout(exam *model.Exam) (Layout, error) {
	if exam.CertificateLayout != nil {
		var layout Layout
		if err := json.Unmarshal(*exam.CertificateLayout, &layout); err != nil {
			return Layout{}, fmt.Errorf("unmarshal certificate layout: %w", err)
		}
		return layout, nil
	}
	tmpl := exam.CertificateTemplate
	if tmpl == "custom" {
		tmpl = "classic"
	}
	return defaultLayout(tmpl), nil
}

// downloadCertificateBackground fetches an uploaded custom background from the
// private bucket by its object key — never a raw or presigned URL is stored
// (FR-18), so every render downloads fresh.
func (s *Service) downloadCertificateBackground(ctx context.Context, key string) ([]byte, error) {
	if s.storage == nil {
		return nil, errors.New("storage not configured")
	}
	obj, err := s.storage.GetObject(ctx, s.cfg.ObjectStorageBucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	// minio-go defers the request until Stat/Read, so a missing object surfaces here.
	if _, err := obj.Stat(); err != nil {
		return nil, err
	}
	return io.ReadAll(obj)
}

// resolveCertificateBackground returns the background image bytes for an
// exam's certificate: the embedded built-in asset for classic/modern/elegant,
// or the downloaded custom upload. A "custom" template with a NULL background
// key falls back to the classic built-in rather than failing (FR-19,
// Invariant 8) — there is no input state where generation fails for lack of a
// template.
func (s *Service) resolveCertificateBackground(ctx context.Context, exam *model.Exam) ([]byte, error) {
	if exam.CertificateTemplate == "custom" {
		if exam.CertificateBackgroundKey != nil {
			return s.downloadCertificateBackground(ctx, *exam.CertificateBackgroundKey)
		}
		return builtinCertificateBackground("classic"), nil
	}
	return builtinCertificateBackground(exam.CertificateTemplate), nil
}

// uploadCertificatePDF uploads a PDF certificate at certificates/<sessionID>.pdf
// and returns its object key. The bucket is private, so callers presign a GET to
// serve it — see resolveCertificateURL.
func (s *Service) uploadCertificatePDF(ctx context.Context, sessionID uuid.UUID, pdf []byte) (string, error) {
	if s.storage == nil {
		return "", errors.New("storage not configured")
	}

	bucket := s.cfg.ObjectStorageBucketName
	key := fmt.Sprintf("certificates/%s.pdf", sessionID.String())
	if _, err := s.storage.PutObject(ctx, bucket, key, bytes.NewReader(pdf), int64(len(pdf)), minio.PutObjectOptions{
		ContentType: "application/pdf",
	}); err != nil {
		return "", err
	}

	return key, nil
}

// latestGradedAt returns the latest non-nil GradedAt across all answers, or nil.
func latestGradedAt(answers []model.ExamSessionAnswer) *time.Time {
	var latest *time.Time
	for _, a := range answers {
		if a.GradedAt != nil {
			if latest == nil || a.GradedAt.After(*latest) {
				latest = a.GradedAt
			}
		}
	}
	return latest
}

// resolveCertificateURL determines a presigned certificate URL for a session,
// regenerating the PDF when missing, stale by grading, or stale by design edit
// (exam.certificate_design_updated_at newer than sess.certificate_generated_at,
// FR-13/C3). The DB stores the object key; a fresh time-limited GET is signed
// on every read since the bucket is private. A non-submitted session resolves
// to (nil, nil) and generates nothing (FR-16); generation/upload/persist
// failures are returned. Regeneration reuses the session's original
// certificate number — AllocateCertificateNumber is idempotent (FR-10).
func (s *Service) resolveCertificateURL(ctx context.Context, exam *model.Exam, sess *model.ExamSession, answers []model.ExamSessionAnswer, studentName string) (*string, error) {
	if sess.Status != "submitted" {
		return nil, nil
	}

	gradedAt := latestGradedAt(answers)
	designStale := exam.CertificateDesignUpdatedAt != nil && sess.CertificateGeneratedAt != nil &&
		exam.CertificateDesignUpdatedAt.After(*sess.CertificateGeneratedAt)

	// Regenerate when certificate is missing, grading is newer, or the design changed.
	if sess.CertificateURL == nil || sess.CertificateGeneratedAt == nil ||
		(gradedAt != nil && gradedAt.After(*sess.CertificateGeneratedAt)) || designStale {

		number, err := s.storeRepo.AllocateCertificateNumber(ctx, sess.ID)
		if err != nil {
			return nil, fmt.Errorf("allocate certificate number: %w", err)
		}
		layout, err := resolveCertificateLayout(exam)
		if err != nil {
			return nil, err
		}
		bg, err := s.resolveCertificateBackground(ctx, exam)
		if err != nil {
			return nil, fmt.Errorf("resolve certificate background: %w", err)
		}

		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return nil, err
		}
		dateStr := sess.SubmittedAt.In(loc).Format("2 January 2006")
		vals := certificateFieldValues(exam.Title, studentName, dateStr, number)

		pdf, err := renderCertificate(layout, bg, vals)
		if err != nil {
			return nil, fmt.Errorf("generate certificate pdf: %w", err)
		}
		key, err := s.uploadCertificatePDF(ctx, sess.ID, pdf)
		if err != nil {
			return nil, fmt.Errorf("upload certificate pdf: %w", err)
		}
		now := time.Now()
		if err := s.storeRepo.UpdateSessionCertificate(ctx, sess.ID, key, now); err != nil {
			return nil, fmt.Errorf("persist certificate key: %w", err)
		}
		sess.CertificateURL = &key
		sess.CertificateNumber = &number
	}

	signed, err := s.presignReadURL(ctx, s.cfg.ObjectStorageBucketName, *sess.CertificateURL, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("presign certificate url: %w", err)
	}
	return &signed, nil
}

// GetCertificatePreview renders a preview certificate through the same
// background/layout resolution as real generation, using a placeholder
// student name ("Nama Peserta Contoh") and placeholder certificate number
// (ABK/2026/000000) — it never allocates a real number (FR-12), since preview
// has no session to allocate against. templateOverride may be empty to use
// the exam's stored template; when it names a different template, the saved
// custom background/layout (authored for the stored template) are not carried
// over, and the override's own built-in default applies instead.
func (s *Service) GetCertificatePreview(ctx context.Context, examID uuid.UUID, templateOverride string) ([]byte, error) {
	exam, err := s.storeRepo.GetExamByID(ctx, examID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrExamNotFound
		}
		return nil, err
	}

	tmpl := templateOverride
	if tmpl == "" {
		tmpl = exam.CertificateTemplate
	}
	if err := validateCertificateTemplate(tmpl); err != nil {
		return nil, err
	}

	previewExam := *exam
	previewExam.CertificateTemplate = tmpl
	if templateOverride != "" && templateOverride != exam.CertificateTemplate {
		previewExam.CertificateBackgroundKey = nil
		previewExam.CertificateLayout = nil
	}

	layout, err := resolveCertificateLayout(&previewExam)
	if err != nil {
		return nil, err
	}
	bg, err := s.resolveCertificateBackground(ctx, &previewExam)
	if err != nil {
		return nil, fmt.Errorf("resolve certificate background: %w", err)
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	vals := certificateFieldValues(exam.Title, "Nama Peserta Contoh", time.Now().In(loc).Format("2 January 2006"), "ABK/2026/000000")

	return renderCertificate(layout, bg, vals)
}
