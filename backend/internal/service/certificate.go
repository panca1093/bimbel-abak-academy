package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

func validateCertificateTemplate(tmpl string) error {
	if !validCertificateTemplates[tmpl] {
		return fmt.Errorf("%w: unknown certificate template: %s", ErrValidation, tmpl)
	}
	return nil
}

// generateCertificatePDF generates a single-page A4 landscape PDF (841.89 x 595.28 pt)
// with the given template, student name, exam title, and submission date (Asia/Jakarta, no score).
// backgroundBytes is optional (nil for built-in templates; used only for custom template in Task 5).
func generateCertificatePDF(template, studentName, examTitle string, submittedAt time.Time, backgroundBytes []byte) ([]byte, error) {
	if err := validateCertificateTemplate(template); err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	dateStr := submittedAt.In(loc).Format("2 January 2006")

	var pdf *gofpdf.Fpdf
	switch template {
	case "classic":
		pdf = classicLayoutGofpdf(studentName, examTitle, dateStr)
	case "modern":
		pdf = modernLayoutGofpdf(studentName, examTitle, dateStr)
	case "elegant":
		pdf = elegantLayoutGofpdf(studentName, examTitle, dateStr)
	case "custom":
		pdf = customLayoutGofpdf(studentName, examTitle, dateStr, backgroundBytes)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// classicLayoutGofpdf renders a gofpdf-based navy-themed certificate with decorative elements.
// Colors from logo-tokens.css: navy-700 (#22315B).
func classicLayoutGofpdf(name, exam, date string) *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "pt", "A4", "")
	pdf.SetCompression(false)
	pdf.AddPage()
	pdf.SetAutoPageBreak(false, 0)

	// Navy color components (0x22315B)
	navyR, navyG, navyB := 0x22, 0x31, 0x5B
	lightBlueR, lightBlueG, lightBlueB := 204, 224, 242 // light blue tint

	// Decorative background: light blue tint
	pdf.SetFillColor(lightBlueR, lightBlueG, lightBlueB)
	pdf.Rect(0, 0, 841.89, 595.28, "F")

	// White interior rectangle
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(50, 50, 741.89, 495.28, "F")

	// Navy border
	pdf.SetDrawColor(navyR, navyG, navyB)
	pdf.SetLineWidth(3)
	pdf.Rect(50, 50, 741.89, 495.28, "D")

	// Navy header band (flush with the page's top edge)
	pdf.SetFillColor(navyR, navyG, navyB)
	pdf.Rect(50, 50, 741.89, 50, "F")

	// Header title (white text on navy band)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 26)
	pdf.SetXY(50, 55)
	pdf.CellFormat(741.89, 30, "CERTIFICATE OF COMPLETION", "", 0, "C", false, 0, "")

	// Circular seal with concentric circles
	drawSeal(pdf, 120, 395, 30, navyR, navyG, navyB)

	// Body text (navy color)
	pdf.SetTextColor(navyR, navyG, navyB)
	pdf.SetFont("Helvetica", "", 14)
	pdf.SetXY(50, 155)
	pdf.CellFormat(741.89, 20, "This certificate is awarded to", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 36)
	pdf.SetXY(50, 195)
	pdf.CellFormat(741.89, 40, name, "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 14)
	pdf.SetXY(50, 265)
	pdf.CellFormat(741.89, 20, "For successfully completing", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetXY(50, 310)
	pdf.CellFormat(741.89, 25, exam, "", 1, "C", false, 0, "")

	// Date
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetXY(50, 375)
	pdf.CellFormat(741.89, 20, fmt.Sprintf("Date: %s", date), "", 1, "C", false, 0, "")

	// Signature line
	pdf.SetLineWidth(1)
	pdf.SetDrawColor(navyR, navyG, navyB)
	pdf.Line(300, 435, 541.89, 435)
	pdf.SetXY(50, 440)
	pdf.CellFormat(741.89, 15, "Authorized Signature", "", 1, "C", false, 0, "")

	return pdf
}

// modernLayoutGofpdf renders a gofpdf-based teal-themed certificate with decorative elements.
// Colors from logo-tokens.css: teal-500 (#1E978A).
func modernLayoutGofpdf(name, exam, date string) *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "pt", "A4", "")
	pdf.SetCompression(false)
	pdf.AddPage()
	pdf.SetAutoPageBreak(false, 0)

	// Teal color components (0x1E978A)
	tealR, tealG, tealB := 0x1E, 0x97, 0x8A

	// White background
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(0, 0, 841.89, 595.28, "F")

	// Teal border
	pdf.SetDrawColor(tealR, tealG, tealB)
	pdf.SetLineWidth(4)
	pdf.Rect(50, 50, 741.89, 495.28, "D")

	// Thin teal accent lines
	pdf.SetLineWidth(2)
	pdf.Line(50, 540, 791.89, 540)
	pdf.Line(50, 55, 791.89, 55)

	// Teal header band (flush with the page's top edge)
	pdf.SetFillColor(tealR, tealG, tealB)
	pdf.Rect(50, 0, 741.89, 50, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 30)
	pdf.SetXY(50, 10)
	pdf.CellFormat(741.89, 30, "CERTIFICATE", "", 0, "C", false, 0, "")

	// Circular seal with concentric circles
	drawSeal(pdf, 120, 395, 30, tealR, tealG, tealB)

	// Body text (teal color)
	pdf.SetTextColor(tealR, tealG, tealB)
	pdf.SetFont("Helvetica", "", 14)
	pdf.SetXY(50, 105)
	pdf.CellFormat(741.89, 20, "Proudly presented to", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 40)
	pdf.SetXY(50, 170)
	pdf.CellFormat(741.89, 45, name, "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 14)
	pdf.SetXY(50, 255)
	pdf.CellFormat(741.89, 20, "For completing the examination", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 24)
	pdf.SetXY(50, 300)
	pdf.CellFormat(741.89, 25, exam, "", 1, "C", false, 0, "")

	// Date
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetXY(50, 375)
	pdf.CellFormat(741.89, 20, fmt.Sprintf("Date: %s", date), "", 1, "C", false, 0, "")

	// Signature line
	pdf.SetLineWidth(1)
	pdf.SetDrawColor(tealR, tealG, tealB)
	pdf.Line(300, 420, 541.89, 420)
	pdf.SetXY(50, 430)
	pdf.CellFormat(741.89, 15, "Authorized Signature", "", 1, "C", false, 0, "")

	return pdf
}

// elegantLayoutGofpdf renders a gofpdf-based gold-themed certificate with decorative elements.
// Colors from logo-tokens.css: gold-600 (#C6881F).
func elegantLayoutGofpdf(name, exam, date string) *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "pt", "A4", "")
	pdf.SetCompression(false)
	pdf.AddPage()
	pdf.SetAutoPageBreak(false, 0)

	// Gold color components (0xC6881F)
	goldR, goldG, goldB := 0xC6, 0x88, 0x1F
	creamR, creamG, creamB := 251, 250, 246 // warm off-white background

	// Cream background
	pdf.SetFillColor(creamR, creamG, creamB)
	pdf.Rect(0, 0, 841.89, 595.28, "F")

	// Gold double border
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(2)
	pdf.Rect(50, 50, 741.89, 495.28, "D")

	pdf.SetLineWidth(1)
	pdf.Rect(55, 55, 731.89, 485.28, "D")

	// Gold horizontal accent lines
	pdf.SetLineWidth(3)
	pdf.Line(50, 95, 791.89, 95)
	pdf.Line(50, 500, 791.89, 500)

	// Circular seal with concentric circles
	drawSeal(pdf, 120, 395, 30, goldR, goldG, goldB)

	// Body text (gold color)
	pdf.SetTextColor(goldR, goldG, goldB)
	pdf.SetFont("Helvetica", "B", 28)
	pdf.SetXY(50, 50)
	pdf.CellFormat(741.89, 25, "Certificate of Achievement", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 12)
	pdf.SetXY(50, 115)
	pdf.CellFormat(741.89, 20, "This certificate is proudly presented to", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 38)
	pdf.SetXY(50, 170)
	pdf.CellFormat(741.89, 40, name, "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 12)
	pdf.SetXY(50, 265)
	pdf.CellFormat(741.89, 20, "For successful completion of", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetXY(50, 310)
	pdf.CellFormat(741.89, 25, exam, "", 1, "C", false, 0, "")

	// Date
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetXY(50, 375)
	pdf.CellFormat(741.89, 20, fmt.Sprintf("Date: %s", date), "", 1, "C", false, 0, "")

	// Signature line
	pdf.SetLineWidth(1)
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.Line(300, 435, 541.89, 435)
	pdf.SetXY(50, 445)
	pdf.CellFormat(741.89, 15, "Authorized Signature", "", 1, "C", false, 0, "")

	return pdf
}

// drawSeal draws a circular seal/emblem at position (x, y) with radius r and the given RGB color.
func drawSeal(pdf *gofpdf.Fpdf, x, y, r float64, r8, g8, b8 int) {
	// Outer circle (filled)
	pdf.SetFillColor(r8, g8, b8)
	pdf.SetDrawColor(r8, g8, b8)
	pdf.Circle(x, y, r, "F")

	// Middle concentric circle (white)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(r8, g8, b8)
	pdf.SetLineWidth(1.5)
	pdf.Circle(x, y, r*0.7, "FD")

	// Inner circle (colored, filled)
	pdf.SetFillColor(r8, g8, b8)
	pdf.Circle(x, y, r*0.4, "F")

	// Central glyph/symbol: white dot in the middle
	pdf.SetFillColor(255, 255, 255)
	pdf.Circle(x, y, r*0.15, "F")
}

// customLayoutGofpdf renders a custom-template certificate with an admin-provided background image.
// The background image (if provided via backgroundBytes) is drawn full-bleed at the page rect.
// Student name, exam title, and date are overlaid at fixed positions.
// No seal or signature line (FR15). If backgroundBytes is empty/nil or image registration fails,
// renders on a plain white page with the text fields (FR14 tolerance).
func customLayoutGofpdf(name, exam, date string, backgroundBytes []byte) *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "pt", "A4", "")
	pdf.SetCompression(false)
	pdf.AddPage()
	pdf.SetAutoPageBreak(false, 0)

	// Page dimensions: A4 landscape is 841.89 x 595.28 pt
	pageWidth := 841.89
	pageHeight := 595.28

	// Try to draw the background image if provided (FR16: full-bleed with explicit width/height).
	// On error or empty bytes, skip the image and render on plain white (FR14).
	if len(backgroundBytes) > 0 {
		// Detect image type from magic bytes; support PNG, JPEG, GIF.
		imgType := detectImageType(backgroundBytes)
		if imgType != "" {
			reader := bytes.NewReader(backgroundBytes)
			// Attempt to register the image. If it succeeds, draw it full-bleed.
			opts := gofpdf.ImageOptions{
				ImageType: imgType,
			}
			info := pdf.RegisterImageOptionsReader("bgimage", opts, reader)
			if info != nil {
				// Image registered successfully; draw it full-bleed covering the entire page.
				pdf.Image("bgimage", 0, 0, pageWidth, pageHeight, false, "", 0, "")
			}
		}
		// If detection fails or registration fails (info == nil), silently skip — render on white instead (FR14).
	}

	// Ensure the background is white if no image was drawn (or image draw failed).
	// If an image was successfully drawn, this won't overwrite it.
	// We'll just ensure text color and positioning are correct.

	// Draw student name in a large font.
	pdf.SetTextColor(0, 0, 0)  // Black text
	pdf.SetFont("Helvetica", "B", 36)
	pdf.SetXY(50, 250)
	pdf.CellFormat(pageWidth-100, 40, name, "", 1, "C", false, 0, "")

	// Draw exam title.
	pdf.SetFont("Helvetica", "", 14)
	pdf.SetXY(50, 300)
	pdf.CellFormat(pageWidth-100, 20, "For successfully completing", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetXY(50, 320)
	pdf.CellFormat(pageWidth-100, 25, exam, "", 1, "C", false, 0, "")

	// Draw submission date.
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetXY(50, 370)
	pdf.CellFormat(pageWidth-100, 20, fmt.Sprintf("Date: %s", date), "", 1, "C", false, 0, "")

	// No seal or signature line per FR15.

	return pdf
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
// regenerating the PDF when missing or stale (latest graded-at is newer than
// generated-at). The DB stores the object key; a fresh time-limited GET is
// signed on every read since the bucket is private. A non-submitted session
// resolves to (nil, nil); generation/upload/persist failures are returned.
func (s *Service) resolveCertificateURL(ctx context.Context, exam *model.Exam, sess *model.ExamSession, answers []model.ExamSessionAnswer, studentName string) (*string, error) {
	if sess.Status != "submitted" {
		return nil, nil
	}

	gradedAt := latestGradedAt(answers)

	// Regenerate when certificate is missing or grading is newer.
	if sess.CertificateURL == nil || sess.CertificateGeneratedAt == nil || (gradedAt != nil && gradedAt.After(*sess.CertificateGeneratedAt)) {
		// Fetch background image if template is "custom" (FR12, FR14).
		var backgroundBytes []byte
		if exam.CertificateTemplate == "custom" && exam.CertificateBackgroundURL != nil && *exam.CertificateBackgroundURL != "" {
			if s.storage != nil {
				reader, err := s.storage.GetObject(ctx, s.cfg.ObjectStorageBucketName, *exam.CertificateBackgroundURL, minio.GetObjectOptions{})
				if err != nil {
					// Tolerate fetch errors (FR14) — proceed with nil background
					reader = nil
				} else {
					// Read all bytes from the object
					var buf bytes.Buffer
					if _, err := buf.ReadFrom(reader); err != nil {
						// Tolerate read errors — proceed with nil background
						backgroundBytes = nil
					} else {
						backgroundBytes = buf.Bytes()
					}
				}
			}
		}

		pdf, err := generateCertificatePDF(exam.CertificateTemplate, studentName, exam.Title, *sess.SubmittedAt, backgroundBytes)
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
	}

	signed, err := s.presignReadURL(ctx, s.cfg.ObjectStorageBucketName, *sess.CertificateURL, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("presign certificate url: %w", err)
	}
	return &signed, nil
}

// GetCertificatePreview generates a preview certificate for admin display using a dummy
// student name ("Nama Peserta Contoh") and the real exam title. templateOverride may be
// empty to use the exam's default template.
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

	// Fetch background image if template is "custom" (FR12, FR14).
	var backgroundBytes []byte
	if tmpl == "custom" && exam.CertificateBackgroundURL != nil && *exam.CertificateBackgroundURL != "" {
		if s.storage != nil {
			reader, err := s.storage.GetObject(ctx, s.cfg.ObjectStorageBucketName, *exam.CertificateBackgroundURL, minio.GetObjectOptions{})
			if err != nil {
				// Tolerate fetch errors (FR14) — proceed with nil background
				reader = nil
			} else {
				// Read all bytes from the object
				var buf bytes.Buffer
				if _, err := buf.ReadFrom(reader); err != nil {
					// Tolerate read errors — proceed with nil background
					backgroundBytes = nil
				} else {
					backgroundBytes = buf.Bytes()
				}
			}
		}
	}

	return generateCertificatePDF(tmpl, "Nama Peserta Contoh", exam.Title, time.Now(), backgroundBytes)
}
