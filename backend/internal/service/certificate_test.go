package service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

// ---------- fakeSessionRepo: certificate extensions ----------

func (f *fakeSessionRepo) UpdateSessionCertificate(_ context.Context, sessionID uuid.UUID, url string, generatedAt time.Time) error {
	s, ok := f.sessions[sessionID]
	if !ok {
		return repository.ErrNotFound
	}
	s.CertificateURL = &url
	s.CertificateGeneratedAt = &generatedAt
	return nil
}

// ---------- shimSessionService: certificate shim ----------

func (s *shimSessionService) uploadCertificatePDF(_ context.Context, sessionID uuid.UUID, _ []byte) (string, error) {
	return "http://minio.example.com/certificates/" + sessionID.String() + ".pdf", nil
}

// resolveCertificateURL mirrors the real Service.resolveCertificateURL using the fake repo
// and a fake upload function — follows the shimSessionService convention from
// exam_session_test.go / exam_result_test.go.
func (s *shimSessionService) resolveCertificateURL(ctx context.Context, exam *model.Exam, sess *model.ExamSession, answers []model.ExamSessionAnswer, studentName string) *string {
	if sess.Status != "submitted" {
		return nil
	}

	gradedAt := latestGradedAt(answers)

	if sess.CertificateURL == nil || sess.CertificateGeneratedAt == nil || (gradedAt != nil && gradedAt.After(*sess.CertificateGeneratedAt)) {
		pdf, err := generateCertificatePDF(exam.CertificateTemplate, studentName, exam.Title, *sess.SubmittedAt)
		if err != nil {
			return nil
		}
		url, err := s.uploadCertificatePDF(ctx, sess.ID, pdf)
		if err != nil {
			return nil
		}
		now := time.Now()
		if err := s.repo.UpdateSessionCertificate(ctx, sess.ID, url, now); err != nil {
			return nil
		}
		return &url
	}

	return sess.CertificateURL
}

// ---------- tests: latestGradedAt ----------

func TestLatestGradedAt_NilWhenEmpty(t *testing.T) {
	t.Parallel()
	got := latestGradedAt(nil)
	if got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

func TestLatestGradedAt_NilWhenAllUngraded(t *testing.T) {
	t.Parallel()
	answers := []model.ExamSessionAnswer{
		{GradedAt: nil},
		{GradedAt: nil},
	}
	got := latestGradedAt(answers)
	if got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

func TestLatestGradedAt_ReturnsMax(t *testing.T) {
	t.Parallel()
	t1 := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	answers := []model.ExamSessionAnswer{
		{GradedAt: &t1},
		{GradedAt: nil},  // ungraded
		{GradedAt: &t2},  // latest
	}
	got := latestGradedAt(answers)
	if got == nil || !got.Equal(t2) {
		t.Errorf("want %v, got %v", t2, got)
	}
}

// ---------- tests: validateCertificateTemplate ----------

func TestValidateCertificateTemplate_ValidKeys(t *testing.T) {
	for _, k := range []string{"classic", "modern", "elegant"} {
		if err := validateCertificateTemplate(k); err != nil {
			t.Errorf("valid key %q: want nil, got %v", k, err)
		}
	}
}

func TestValidateCertificateTemplate_InvalidKey(t *testing.T) {
	err := validateCertificateTemplate("unknown")
	if !errors.Is(err, ErrValidation) {
		t.Errorf("want ErrValidation, got %v", err)
	}
}

// ---------- tests: generateCertificatePDF ----------

func TestGenerateCertificatePDF_DifferentTemplates_Distinguishable(t *testing.T) {
	now := time.Now()
	classicPDF, err := generateCertificatePDF("classic", "Budi", "Test Exam", now)
	if err != nil {
		t.Fatalf("classic: %v", err)
	}
	modernPDF, err := generateCertificatePDF("modern", "Budi", "Test Exam", now)
	if err != nil {
		t.Fatalf("modern: %v", err)
	}
	if bytes.Equal(classicPDF, modernPDF) {
		t.Error("classic and modern should produce different PDF bytes")
	}

	elegantPDF, err := generateCertificatePDF("elegant", "Budi", "Test Exam", now)
	if err != nil {
		t.Fatalf("elegant: %v", err)
	}
	if bytes.Equal(elegantPDF, classicPDF) {
		t.Error("elegant and classic should produce different PDF bytes")
	}
	if bytes.Equal(elegantPDF, modernPDF) {
		t.Error("elegant and modern should produce different PDF bytes")
	}
}

func TestGenerateCertificatePDF_InvalidTemplate(t *testing.T) {
	_, err := generateCertificatePDF("unknown", "Budi", "Test", time.Now())
	if !errors.Is(err, ErrValidation) {
		t.Errorf("want ErrValidation, got %v", err)
	}
}

// ---------- tests: resolveCertificateURL ----------

func TestResolveCertificateURL_NotSubmitted(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	sess := &model.ExamSession{
		ID: uuid.New(), Status: "in_progress",
		SubmittedAt: nil, CertificateURL: nil, CertificateGeneratedAt: nil,
	}
	svc.repo.sessions[sess.ID] = sess
	exam := &model.Exam{CertificateTemplate: "classic", Title: "Test"}

	url := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi")
	if url != nil {
		t.Errorf("want nil for non-submitted session, got %q", *url)
	}
	// No side effects on an in_progress session.
	if svc.repo.sessions[sess.ID].CertificateURL != nil {
		t.Error("CertificateURL should remain nil")
	}
}

func TestResolveCertificateURL_FirstTimeGeneration(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	sess := &model.ExamSession{
		ID: uuid.New(), Status: "submitted", SubmittedAt: &submittedAt,
		CertificateURL: nil, CertificateGeneratedAt: nil,
	}
	svc.repo.sessions[sess.ID] = sess
	exam := &model.Exam{CertificateTemplate: "classic", Title: "My Exam"}
	wantURL := "http://minio.example.com/certificates/" + sess.ID.String() + ".pdf"

	url := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi")
	if url == nil {
		t.Fatal("want non-nil URL for first-time generation")
	}
	if *url != wantURL {
		t.Errorf("URL: want %q, got %q", wantURL, *url)
	}
	// Session was updated.
	updated := svc.repo.sessions[sess.ID]
	if updated.CertificateURL == nil || *updated.CertificateURL != wantURL {
		t.Error("session CertificateURL should be set")
	}
	if updated.CertificateGeneratedAt == nil {
		t.Error("session CertificateGeneratedAt should be set")
	}
}

func TestResolveCertificateURL_NoRegenerationWhenNotStale(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	certGeneratedAt := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	oldURL := "http://old.url/cert.pdf"
	sess := &model.ExamSession{
		ID: uuid.New(), Status: "submitted", SubmittedAt: &submittedAt,
		CertificateURL: &oldURL, CertificateGeneratedAt: &certGeneratedAt,
	}
	svc.repo.sessions[sess.ID] = sess
	exam := &model.Exam{CertificateTemplate: "classic", Title: "My Exam"}

	// Answers graded before the certificate was generated.
	gradedAt := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	answers := []model.ExamSessionAnswer{{GradedAt: &gradedAt}}

	url := svc.resolveCertificateURL(ctx, exam, sess, answers, "Budi")
	if url == nil {
		t.Fatal("want non-nil URL (existing)")
	}
	if *url != oldURL {
		t.Errorf("want existing URL %q, got %q", oldURL, *url)
	}
	// No regeneration occurred — session fields unchanged.
	updated := svc.repo.sessions[sess.ID]
	if updated.CertificateURL == nil || *updated.CertificateURL != oldURL {
		t.Error("session CertificateURL should still be the old URL")
	}
}

func TestResolveCertificateURL_RegenerationWhenStale(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	certGeneratedAt := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	oldURL := "http://old.url/cert.pdf"
	sess := &model.ExamSession{
		ID: uuid.New(), Status: "submitted", SubmittedAt: &submittedAt,
		CertificateURL: &oldURL, CertificateGeneratedAt: &certGeneratedAt,
	}
	svc.repo.sessions[sess.ID] = sess
	exam := &model.Exam{CertificateTemplate: "classic", Title: "My Exam"}

	// Answer graded AFTER certificate was generated → stale → regen.
	gradedAt := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	answers := []model.ExamSessionAnswer{{GradedAt: &gradedAt}}

	url := svc.resolveCertificateURL(ctx, exam, sess, answers, "Budi")
	if url == nil {
		t.Fatal("want non-nil URL (regenerated)")
	}
	if *url == oldURL {
		t.Error("regeneration should produce a different URL")
	}
	if *url != "http://minio.example.com/certificates/"+sess.ID.String()+".pdf" {
		t.Errorf("unexpected URL: %q", *url)
	}
	// Session was updated.
	updated := svc.repo.sessions[sess.ID]
	if updated.CertificateURL == nil || *updated.CertificateURL == oldURL {
		t.Error("session CertificateURL should have been updated")
	}
	if updated.CertificateGeneratedAt == nil {
		t.Error("session CertificateGeneratedAt should be set")
	}
}
