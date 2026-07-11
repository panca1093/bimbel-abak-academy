package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// validCertificateTemplates is the closed set of accepted certificate layouts.
var validCertificateTemplates = map[string]bool{
	"classic": true,
	"modern":  true,
	"elegant": true,
}

func validateCertificateTemplate(tmpl string) error {
	if !validCertificateTemplates[tmpl] {
		return fmt.Errorf("%w: unknown certificate template: %s", ErrValidation, tmpl)
	}
	return nil
}

// pdfEscape escapes (, ), and \ in PDF literal string content.
func pdfEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	return s
}

// generateCertificatePDF generates a single-page A4 landscape PDF (841.89 x 595.28 pt)
// with the given template, student name, exam title, and submission date
// (Asia/Jakarta, no score).
func generateCertificatePDF(template, studentName, examTitle string, submittedAt time.Time) ([]byte, error) {
	if err := validateCertificateTemplate(template); err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	dateStr := pdfEscape(submittedAt.In(loc).Format("2 January 2006"))
	name := pdfEscape(studentName)
	exam := pdfEscape(examTitle)

	var content []byte
	switch template {
	case "classic":
		content = classicLayout(name, exam, dateStr)
	case "modern":
		content = modernLayout(name, exam, dateStr)
	case "elegant":
		content = elegantLayout(name, exam, dateStr)
	}
	return buildPDF(content), nil
}

// classicLayout renders a blue-themed certificate: light-blue background, dark-blue
// header band, blue bordered white interior.
func classicLayout(name, exam, date string) []byte {
	return []byte(fmt.Sprintf(`q
0.8 0.88 0.95 rg
0 0 841.89 595.28 re f
1 1 1 rg
50 50 741.89 495.28 re f
0.2 0.35 0.65 RG
3 w
50 50 741.89 495.28 re S
0.2 0.35 0.65 rg
50 495 741.89 50 re f
1 1 1 rg
BT
/F1 26 Tf
421 525 Td
(CERTIFICATE OF COMPLETION) Tj
ET
0.2 0.35 0.65 rg
BT
/F1 14 Tf
421 450 Td
(This certificate is awarded to) Tj
ET
BT
/F1 36 Tf
421 390 Td
(%s) Tj
ET
BT
/F1 14 Tf
421 320 Td
(For successfully completing) Tj
ET
BT
/F1 22 Tf
421 270 Td
(%s) Tj
ET
BT
/F1 12 Tf
421 200 Td
(Date: %s) Tj
ET
Q`, name, exam, date))
}

// modernLayout renders a teal-themed certificate: white background, teal thick border,
// teal header band with "CERTIFICATE" title.
func modernLayout(name, exam, date string) []byte {
	return []byte(fmt.Sprintf(`q
1 1 1 rg
0 0 841.89 595.28 re f
0 0.44 0.44 RG
4 w
50 50 741.89 495.28 re S
0 0.44 0.44 rg
50 545 741.89 50 re f
0 0.44 0.44 rg
50 50 741.89 2 re f
1 1 1 rg
BT
/F1 30 Tf
421 570 Td
(CERTIFICATE) Tj
ET
BT
/F1 14 Tf
421 490 Td
(Proudly presented to) Tj
ET
0 0.44 0.44 rg
BT
/F1 40 Tf
421 420 Td
(%s) Tj
ET
BT
/F1 14 Tf
421 340 Td
(For completing the examination) Tj
ET
0 0.44 0.44 rg
BT
/F1 24 Tf
421 280 Td
(%s) Tj
ET
BT
/F1 12 Tf
421 200 Td
(Date: %s) Tj
ET
Q`, name, exam, date))
}

// elegantLayout renders a gold-toned certificate: cream background, gold double border,
// gold accent lines.
func elegantLayout(name, exam, date string) []byte {
	return []byte(fmt.Sprintf(`q
1 0.98 0.9 rg
0 0 841.89 595.28 re f
0.55 0.42 0.12 RG
2 w
50 50 741.89 495.28 re S
0.55 0.42 0.12 RG
1 w
55 55 731.89 485.28 re S
0.55 0.42 0.12 rg
50 500 741.89 3 re f
0.55 0.42 0.12 rg
50 75 741.89 3 re f
BT
/F1 28 Tf
421 540 Td
(Certificate of Achievement) Tj
ET
BT
/F1 12 Tf
421 475 Td
(This certificate is proudly presented to) Tj
ET
BT
/F1 38 Tf
421 400 Td
(%s) Tj
ET
BT
/F1 12 Tf
421 325 Td
(For successful completion of) Tj
ET
BT
/F1 22 Tf
421 270 Td
(%s) Tj
ET
BT
/F1 11 Tf
421 200 Td
(Date: %s) Tj
ET
Q`, name, exam, date))
}

// buildPDF wraps a content stream with the full %PDF-1.4 structure: catalog, pages,
// page (A4 landscape), content stream, Helvetica font, xref table, trailer, startxref, %%EOF.
func buildPDF(content []byte) []byte {
	var buf bytes.Buffer

	// Header
	buf.WriteString("%PDF-1.4\n")

	type objRef struct {
		num    int
		offset int
	}
	var objs []objRef

	// 1 0 obj — Catalog
	objs = append(objs, objRef{1, buf.Len()})
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// 2 0 obj — Pages
	objs = append(objs, objRef{2, buf.Len()})
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// 3 0 obj — Page (A4 landscape)
	objs = append(objs, objRef{3, buf.Len()})
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 841.89 595.28] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n")

	// 4 0 obj — Content stream
	objs = append(objs, objRef{4, buf.Len()})
	buf.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n", len(content)))
	buf.Write(content)
	buf.WriteString("\nendstream\nendobj\n")

	// 5 0 obj — Font (base-14 Helvetica)
	objs = append(objs, objRef{5, buf.Len()})
	buf.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	// xref table
	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString(fmt.Sprintf("0 %d\n", len(objs)+1))
	buf.WriteString("0000000000 65535 f \n")
	for _, o := range objs {
		fmt.Fprintf(&buf, "%010d 00000 n \n", o.offset)
	}

	// Trailer
	buf.WriteString("trailer\n")
	buf.WriteString(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", len(objs)+1))
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefOffset))
	buf.WriteString("%%EOF\n")

	return buf.Bytes()
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
		pdf, err := generateCertificatePDF(exam.CertificateTemplate, studentName, exam.Title, *sess.SubmittedAt)
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

	return generateCertificatePDF(tmpl, "Nama Peserta Contoh", exam.Title, time.Now())
}
