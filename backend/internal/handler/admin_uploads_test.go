package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"akademi-bimbel/internal/infra"
	"github.com/labstack/echo/v4"
)

// setAdminExamClaims sets admin_exam claims on the echo context.
func setAdminExamClaims(c echo.Context, sub string) {
	c.Set("claims", &infra.Claims{Sub: sub, Role: "admin_exam"})
}

// TestAdminUploadImage_ValidImageContentType_PassesValidation verifies that a valid
// image content type passes validation (proceeds to service call).
func TestAdminUploadImage_ValidImageContentType_PassesValidation(t *testing.T) {
	env := newAdminSystemEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads/image?filename=test.png&content_type=image/png", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminExamClaims(c, "exam-admin-1")

	if err := env.h.AdminUploadImage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Validation passes; service call fails (no storage configured in test).
	// We're testing that validation didn't reject it (would be 400 if rejected).
	if rec.Code == http.StatusBadRequest {
		t.Errorf("content-type validation incorrectly rejected image/png")
	}
}

// TestAdminUploadImage_InvalidContentType_Returns400 verifies that non-image
// content types are rejected.
func TestAdminUploadImage_InvalidContentType_Returns400(t *testing.T) {
	env := newAdminSystemEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads/image?filename=test.mp3&content_type=audio/mpeg", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminExamClaims(c, "exam-admin-1")

	if err := env.h.AdminUploadImage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

// TestAdminUploadImage_MissingFilename_Returns400 verifies validation of required params.
func TestAdminUploadImage_MissingFilename_Returns400(t *testing.T) {
	env := newAdminSystemEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads/image?content_type=image/png", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminExamClaims(c, "exam-admin-1")

	if err := env.h.AdminUploadImage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 for missing filename, got %d", rec.Code)
	}
}

// TestAdminUploadImage_NoAuth_Returns403 verifies that unauthenticated requests fail.
func TestAdminUploadImage_NoAuth_Returns403(t *testing.T) {
	env := newAdminSystemEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads/image?filename=test.png&content_type=image/png", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	// No claims set

	if err := env.h.AdminUploadImage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403 for no auth, got %d", rec.Code)
	}
}

// TestAdminUploadImage_WithClaims_ProceedsToService verifies that valid
// claims proceed through handler validation (RBAC check is in middleware).
func TestAdminUploadImage_WithClaims_ProceedsToService(t *testing.T) {
	env := newAdminSystemEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads/image?filename=test.png&content_type=image/png", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	c.Set("claims", &infra.Claims{Sub: "user-1", Role: "student"})

	if err := env.h.AdminUploadImage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Handler proceeds (RBAC is enforced in middleware via routes.go).
	// Service call fails (no storage), but validation passed.
	if rec.Code == http.StatusBadRequest {
		t.Errorf("handler validation incorrectly rejected valid request")
	}
}

// TestAdminUploadAudio_ValidAudioContentType_PassesValidation verifies that a valid
// audio content type passes validation (proceeds to service call).
func TestAdminUploadAudio_ValidAudioContentType_PassesValidation(t *testing.T) {
	env := newAdminSystemEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads/audio?filename=test.mp3&content_type=audio/mpeg", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminExamClaims(c, "exam-admin-1")

	if err := env.h.AdminUploadAudio(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Validation passes; service call fails (no storage configured in test).
	// We're testing that validation didn't reject it (would be 400 if rejected).
	if rec.Code == http.StatusBadRequest {
		t.Errorf("content-type validation incorrectly rejected audio/mpeg")
	}
}

// TestAdminUploadAudio_InvalidContentType_Returns400 verifies that non-audio
// content types are rejected.
func TestAdminUploadAudio_InvalidContentType_Returns400(t *testing.T) {
	env := newAdminSystemEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads/audio?filename=test.png&content_type=image/png", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminExamClaims(c, "exam-admin-1")

	if err := env.h.AdminUploadAudio(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

// TestAdminUploadAudio_NoAuth_Returns403 verifies that unauthenticated requests fail.
func TestAdminUploadAudio_NoAuth_Returns403(t *testing.T) {
	env := newAdminSystemEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/uploads/audio?filename=test.mp3&content_type=audio/mpeg", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	// No claims set

	if err := env.h.AdminUploadAudio(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403 for no auth, got %d", rec.Code)
	}
}
