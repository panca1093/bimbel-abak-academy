package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
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

// AllocateCertificateNumber fakes the repository's idempotent allocation
// (FR-9/FR-10): mints once per session, keyed off the session id so it needs
// no extra counter state on fakeSessionRepo, and returns the same value on
// every later call for that session.
func (f *fakeSessionRepo) AllocateCertificateNumber(_ context.Context, sessionID uuid.UUID) (string, error) {
	s, ok := f.sessions[sessionID]
	if !ok {
		return "", repository.ErrNotFound
	}
	if s.CertificateNumber != nil {
		return *s.CertificateNumber, nil
	}
	number := "ABK/2026/" + sessionID.String()[:6]
	s.CertificateNumber = &number
	return number, nil
}

// ---------- shimSessionService: certificate shim ----------

func (s *shimSessionService) uploadCertificatePDF(_ context.Context, sessionID uuid.UUID, _ []byte) (string, error) {
	s.uploadCertCalls++
	if s.uploadCertErr != nil {
		return "", s.uploadCertErr
	}
	return "http://minio.example.com/certificates/" + sessionID.String() + ".pdf", nil
}

// downloadCertificateBackground fakes the private-bucket download for a custom
// background: returns a real embedded PNG (the classic built-in bytes stand in
// for "whatever was uploaded") so renderCertificate can embed it for real.
func (s *shimSessionService) downloadCertificateBackground(_ context.Context, _ string) ([]byte, error) {
	return certBgClassicPNG, nil
}

// resolveCertificateBackground mirrors the real Service.resolveCertificateBackground:
// built-in templates use the embedded asset; "custom" downloads by key, or falls
// back to classic when the key is NULL (FR-19).
func (s *shimSessionService) resolveCertificateBackground(ctx context.Context, exam *model.Exam) ([]byte, error) {
	if exam.CertificateTemplate == "custom" {
		if exam.CertificateBackgroundKey != nil {
			return s.downloadCertificateBackground(ctx, *exam.CertificateBackgroundKey)
		}
		return builtinCertificateBackground("classic"), nil
	}
	return builtinCertificateBackground(exam.CertificateTemplate), nil
}

// resolveCertificateURL mirrors the real Service.resolveCertificateURL using the fake repo
// and fake I/O boundaries — follows the shimSessionService convention from
// exam_session_test.go / exam_result_test.go. resolveCertificateLayout and
// renderCertificate are pure package functions, so this calls them for real
// rather than faking them.
func (s *shimSessionService) resolveCertificateURL(ctx context.Context, exam *model.Exam, sess *model.ExamSession, answers []model.ExamSessionAnswer, studentName string) (*string, error) {
	if sess.Status != "submitted" {
		return nil, nil
	}

	gradedAt := latestGradedAt(answers)
	designStale := exam.CertificateDesignUpdatedAt != nil && sess.CertificateGeneratedAt != nil &&
		exam.CertificateDesignUpdatedAt.After(*sess.CertificateGeneratedAt)

	if sess.CertificateURL == nil || sess.CertificateGeneratedAt == nil ||
		(gradedAt != nil && gradedAt.After(*sess.CertificateGeneratedAt)) || designStale {

		number, err := s.repo.AllocateCertificateNumber(ctx, sess.ID)
		if err != nil {
			return nil, err
		}
		layout, err := resolveCertificateLayout(exam)
		if err != nil {
			return nil, err
		}
		bg, err := s.resolveCertificateBackground(ctx, exam)
		if err != nil {
			return nil, err
		}

		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return nil, err
		}
		dateStr := sess.SubmittedAt.In(loc).Format("2 January 2006")
		vals := certificateFieldValues(exam.Title, studentName, dateStr, number)

		pdf, err := renderCertificate(layout, bg, vals)
		if err != nil {
			return nil, err
		}
		url, err := s.uploadCertificatePDF(ctx, sess.ID, pdf)
		if err != nil {
			return nil, err
		}
		now := time.Now()
		if err := s.repo.UpdateSessionCertificate(ctx, sess.ID, url, now); err != nil {
			return nil, err
		}
		sess.CertificateNumber = &number
		return &url, nil
	}

	return sess.CertificateURL, nil
}

// GetCertificatePreview mirrors the real Service.GetCertificatePreview: no
// allocation (FR-12), placeholder name/number, same background/layout
// resolution as real generation.
func (s *shimSessionService) GetCertificatePreview(ctx context.Context, examID uuid.UUID, templateOverride string) ([]byte, error) {
	exam, err := s.repo.GetExamForSession(ctx, examID)
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
		return nil, err
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	vals := certificateFieldValues(exam.Title, "Nama Peserta Contoh", time.Now().In(loc).Format("2 January 2006"), "ABK/2026/000000")

	return renderCertificate(layout, bg, vals)
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

	url, err := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != nil {
		t.Errorf("want nil for non-submitted session, got %q", *url)
	}
	// No side effects on an in_progress session.
	if svc.repo.sessions[sess.ID].CertificateURL != nil {
		t.Error("CertificateURL should remain nil")
	}
	if svc.uploadCertCalls != 0 {
		t.Errorf("non-submitted session must generate nothing, got %d upload calls", svc.uploadCertCalls)
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

	url, err := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

	url, err := svc.resolveCertificateURL(ctx, exam, sess, answers, "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

	url, err := svc.resolveCertificateURL(ctx, exam, sess, answers, "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

func TestResolveCertificateURL_UploadFailure_ReturnsError(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)
	svc.uploadCertErr = errors.New("minio down")

	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	sess := &model.ExamSession{
		ID: uuid.New(), Status: "submitted", SubmittedAt: &submittedAt,
		CertificateURL: nil, CertificateGeneratedAt: nil,
	}
	svc.repo.sessions[sess.ID] = sess
	exam := &model.Exam{CertificateTemplate: "classic", Title: "My Exam"}

	url, err := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi")
	if err == nil {
		t.Fatal("want error when upload fails, got nil")
	}
	if url != nil {
		t.Errorf("want nil URL on upload failure, got %q", *url)
	}
	if svc.repo.sessions[sess.ID].CertificateURL != nil {
		t.Error("must not persist a certificate URL when upload failed")
	}
}

func TestResolveCertificateURL_PersistFailure_ReturnsError(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	// Session NOT seeded in the repo → UpdateSessionCertificate returns ErrNotFound.
	sess := &model.ExamSession{
		ID: uuid.New(), Status: "submitted", SubmittedAt: &submittedAt,
		CertificateURL: nil, CertificateGeneratedAt: nil,
	}
	exam := &model.Exam{CertificateTemplate: "classic", Title: "My Exam"}

	url, err := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("want ErrNotFound from persist step, got %v", err)
	}
	if url != nil {
		t.Errorf("want nil URL on persist failure, got %q", *url)
	}
}

// ---------- tests: resolveCertificateURL — design staleness (FR-13/FR-15) ----------

func TestResolveCertificateURL_RegenerationWhenDesignStale(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	certGeneratedAt := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	designUpdatedAt := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC) // after cert generated
	oldURL := "http://old.url/cert.pdf"
	sess := &model.ExamSession{
		ID: uuid.New(), Status: "submitted", SubmittedAt: &submittedAt,
		CertificateURL: &oldURL, CertificateGeneratedAt: &certGeneratedAt,
	}
	svc.repo.sessions[sess.ID] = sess
	exam := &model.Exam{CertificateTemplate: "classic", Title: "My Exam", CertificateDesignUpdatedAt: &designUpdatedAt}

	url, err := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == nil {
		t.Fatal("want non-nil URL")
	}
	if svc.uploadCertCalls != 1 {
		t.Errorf("a design edit after generation should trigger exactly one regeneration, got %d upload calls", svc.uploadCertCalls)
	}
	updated := svc.repo.sessions[sess.ID]
	if updated.CertificateGeneratedAt == nil || !updated.CertificateGeneratedAt.After(certGeneratedAt) {
		t.Error("CertificateGeneratedAt should have been bumped by the regeneration")
	}
}

func TestResolveCertificateURL_NoRegenerationWhenDesignNotStaleOrChanged(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	certGeneratedAt := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	designUpdatedAt := time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC) // before cert generated
	oldURL := "http://old.url/cert.pdf"
	sess := &model.ExamSession{
		ID: uuid.New(), Status: "submitted", SubmittedAt: &submittedAt,
		CertificateURL: &oldURL, CertificateGeneratedAt: &certGeneratedAt,
	}
	svc.repo.sessions[sess.ID] = sess
	exam := &model.Exam{CertificateTemplate: "classic", Title: "My Exam", CertificateDesignUpdatedAt: &designUpdatedAt}

	url, err := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == nil || *url != oldURL {
		t.Errorf("want existing URL %q, got %v", oldURL, url)
	}
	if svc.uploadCertCalls != 0 {
		t.Errorf("want zero regenerations, got %d upload calls", svc.uploadCertCalls)
	}

	// FR-15: a second read with nothing changed must trigger zero further regeneration.
	if _, err := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi"); err != nil {
		t.Fatalf("unexpected error on second read: %v", err)
	}
	if svc.uploadCertCalls != 0 {
		t.Errorf("second read with nothing changed should not regenerate, got %d upload calls", svc.uploadCertCalls)
	}
}

// ---------- tests: resolveCertificateURL — certificate number immutability (FR-10) ----------

func TestResolveCertificateURL_RegenerationReusesCertificateNumber(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	sess := &model.ExamSession{
		ID: uuid.New(), Status: "submitted", SubmittedAt: &submittedAt,
	}
	svc.repo.sessions[sess.ID] = sess
	exam := &model.Exam{CertificateTemplate: "classic", Title: "My Exam"}

	if _, err := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi"); err != nil {
		t.Fatalf("first generation: %v", err)
	}
	firstNumber := svc.repo.sessions[sess.ID].CertificateNumber
	if firstNumber == nil {
		t.Fatal("want a certificate number allocated on first generation")
	}

	// Force a regeneration via re-grading staleness.
	gradedAt := time.Now().Add(time.Hour)
	answers := []model.ExamSessionAnswer{{GradedAt: &gradedAt}}
	if _, err := svc.resolveCertificateURL(ctx, exam, sess, answers, "Budi"); err != nil {
		t.Fatalf("regeneration: %v", err)
	}
	secondNumber := svc.repo.sessions[sess.ID].CertificateNumber
	if secondNumber == nil || *secondNumber != *firstNumber {
		t.Errorf("regeneration should reuse the original number: first=%v second=%v", firstNumber, secondNumber)
	}
	if svc.uploadCertCalls != 2 {
		t.Errorf("want 2 uploads (first generation + regeneration), got %d", svc.uploadCertCalls)
	}
}

// ---------- tests: custom template with NULL background key (FR-19) ----------

func TestResolveCertificateURL_CustomTemplateNilBackgroundKey_Renders(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	submittedAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	sess := &model.ExamSession{
		ID: uuid.New(), Status: "submitted", SubmittedAt: &submittedAt,
	}
	svc.repo.sessions[sess.ID] = sess
	exam := &model.Exam{CertificateTemplate: "custom", CertificateBackgroundKey: nil, Title: "Custom Exam"}

	url, err := svc.resolveCertificateURL(ctx, exam, sess, nil, "Budi")
	if err != nil {
		t.Fatalf("custom template with a NULL background key should still render, got error: %v", err)
	}
	if url == nil {
		t.Fatal("want non-nil URL")
	}
}

// ---------- tests: GetCertificatePreview (FR-12, FR-19) ----------

func TestGetCertificatePreview_DoesNotAllocate(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	exam := &model.Exam{CertificateTemplate: "classic", Title: "Preview Exam"}
	svc.repo.seedExam(exam)

	pdf, err := svc.GetCertificatePreview(ctx, exam.ID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("want a non-empty PDF")
	}
	// No session exists for this exam. If GetCertificatePreview ever called
	// AllocateCertificateNumber, there would be nothing to allocate against —
	// the fake repo would return ErrNotFound and this call would fail.
	if len(svc.repo.sessions) != 0 {
		t.Fatal("preview must not create or touch any session")
	}
}

func TestGetCertificatePreview_CustomTemplateNilBackgroundKey_Renders(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	exam := &model.Exam{CertificateTemplate: "custom", CertificateBackgroundKey: nil, Title: "Custom Preview Exam"}
	svc.repo.seedExam(exam)

	pdf, err := svc.GetCertificatePreview(ctx, exam.ID, "")
	if err != nil {
		t.Fatalf("custom template with a NULL background key should still render, got error: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("want a non-empty PDF")
	}
}

func TestGetCertificatePreview_UnknownExam_ReturnsErrExamNotFound(t *testing.T) {
	ctx := context.Background()
	svc, _ := newShimSessionService(t)

	_, err := svc.GetCertificatePreview(ctx, uuid.New(), "")
	if !errors.Is(err, ErrExamNotFound) {
		t.Errorf("want ErrExamNotFound, got %v", err)
	}
}

// ---------- tests: resolveCertificateLayout (FR-29) ----------

func TestResolveCertificateLayout_NilLayout_SeedsBuiltinDefault(t *testing.T) {
	exam := &model.Exam{CertificateTemplate: "modern", CertificateLayout: nil}
	layout, err := resolveCertificateLayout(exam)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := defaultLayout("modern")
	if !reflect.DeepEqual(layout, want) {
		t.Error("an exam with no saved layout should seed the built-in template's default layout, not an empty canvas")
	}
}

func TestResolveCertificateLayout_CustomTemplateNilLayout_FallsBackToClassic(t *testing.T) {
	exam := &model.Exam{CertificateTemplate: "custom", CertificateLayout: nil}
	layout, err := resolveCertificateLayout(exam)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := defaultLayout("classic")
	if !reflect.DeepEqual(layout, want) {
		t.Error("a custom template with no saved layout should fall back to classic's default layout")
	}
}

func TestResolveCertificateLayout_SavedLayout_UsesPersistedFields(t *testing.T) {
	raw := json.RawMessage(`{"page":{"width_mm":297,"height_mm":210},"background":{"kind":"builtin","ref":"classic"},"fields":[{"id":"title","x_mm":10,"y_mm":10,"w_mm":50,"align":"left","visible":true}]}`)
	exam := &model.Exam{CertificateTemplate: "classic", CertificateLayout: &raw}
	layout, err := resolveCertificateLayout(exam)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(layout.Fields) != 1 || layout.Fields[0].ID != "title" || layout.Fields[0].XMm != 10 {
		t.Errorf("want the persisted layout fields, got %+v", layout.Fields)
	}
}
